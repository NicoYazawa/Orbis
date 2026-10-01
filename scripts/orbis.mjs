import { execFileSync, spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertPortsAvailable } from './ports.mjs';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const argv = process.argv.slice(2);
const command = argv.shift() ?? 'start';
const profileArg = argv.find((arg) => arg.startsWith('--profiles='));
const profiles = (profileArg?.slice(11) ?? 'app,rag').split(',').filter(Boolean);
const valid = new Set(['app','rag','obs','ollama','search']);
for (const profile of profiles) if (!valid.has(profile)) throw new Error(`Unsupported profile: ${profile}`);
const dev = argv.includes('--dev');
if (profiles.includes('obs') && !process.env.ORBIS_OTEL_ENDPOINT) process.env.ORBIS_OTEL_ENDPOINT = 'http://jaeger:4318';
const project = 'orbis';
const args = ['compose', '-p', project, '--env-file', path.join(root,'.env'), '-f', path.join(root,'compose.yaml')];
if (dev) args.push('-f',path.join(root,'compose.dev.yaml'));
for (const profile of profiles) args.push('--profile', profile);

function capture(extra) {
  try { return execFileSync('docker',[...args,...extra],{cwd:root,encoding:'utf8',stdio:['ignore','pipe','pipe'],maxBuffer:8*1024*1024}); }
  catch { throw new Error(`Docker ${extra[0]} failed. Check Docker Desktop, .env and Compose configuration; raw configuration is withheld to protect credentials.`); }
}
function execute(extra) {
  const result=spawnSync('docker',[...args,...extra],{cwd:root,stdio:'inherit'});
  if (result.error) throw new Error(`Docker unavailable: ${result.error.code}`);
  if(result.status!==0) throw new Error(`Docker ${extra[0]} failed (${result.status}); other projects were not changed.`);
}
function initialize() {
  if (fs.existsSync(path.join(root,'.env'))) return;
  let template=fs.readFileSync(path.join(root,'.env.example'),'utf8');
  for (const key of ['ORBIS_POSTGRES_PASSWORD','ORBIS_S3_ACCESS_KEY','ORBIS_S3_SECRET_KEY','ORBIS_ADMIN_PASSWORD','ORBIS_GRAFANA_PASSWORD']) {
    template=template.replace(new RegExp(`^${key}=$`,'m'),`${key}=${randomBytes(24).toString('hex')}`);
  }
  fs.writeFileSync(path.join(root,'.env'),template,{mode:0o600,flag:'wx'});
  console.log('Created local .env with random credentials; values were not printed. Do not commit it.');
}
async function preflight() {
  const config=JSON.parse(capture(['config','--format','json']));
  const selected=new Set(Object.entries(config.services).filter(([,service])=>!service.profiles || service.profiles.some((p)=>profiles.includes(p))).map(([name])=>name));
  // Includes transitive dependencies selected by Compose.
  for(const name of selected) for(const dependency of Object.keys(config.services[name].depends_on??{})) selected.add(dependency);
  const mappings=[];
  for(const name of selected) for(const port of config.services[name].ports??[]) mappings.push({service:name,host:port.host_ip??'0.0.0.0',port:Number(port.published),protocol:port.protocol??'tcp'});
  const owned=new Set();
  const ids=capture(['ps','-q']).trim().split(/\s+/).filter(Boolean);
  if(ids.length) {
    const inspectFormat='{"project":{{json (index .Config.Labels "com.docker.compose.project")}},"service":{{json (index .Config.Labels "com.docker.compose.service")}},"ports":{{json .NetworkSettings.Ports}}}';
    const existing=execFileSync('docker',['inspect','--format',inspectFormat,...ids],{cwd:root,encoding:'utf8',maxBuffer:1024*1024}).trim().split('\n').filter(Boolean).map((line)=>JSON.parse(line));
    for(const container of existing) {
      if(container.project!==project) continue;
      const service=container.service;
      for(const [target,bindings] of Object.entries(container.ports??{})) for(const binding of bindings??[]) {
        if(binding.HostIp==='127.0.0.1') owned.add(`${service}:${binding.HostPort}/${target.split('/')[1]}`);
      }
    }
  }
  await assertPortsAvailable(mappings,owned);
  console.log(`Port preflight passed: ${mappings.map((p)=>`${p.service}=${p.host}:${p.port}`).join(', ') || 'internal services only'}`);
}

try {
  if(command==='init') initialize();
  else if(command==='start'||command==='check-ports') {
    initialize();
    await preflight();
    if(command==='start') {
      const up=['up','-d','--wait','--wait-timeout','180'];
      if(!argv.includes('--no-build')) up.push('--build');
      execute(up);
    }
  } else if(command==='stop') execute(['down','--remove-orphans']);
  else if(command==='status') execute(['ps','-a']);
  else if(command==='verify') execute(['run','--rm','--no-deps','verify']);
  else throw new Error(`Unknown command: ${command}`);
} catch(error) { console.error(error.message); process.exitCode=1; }
