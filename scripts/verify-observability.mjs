import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const config=fs.readFileSync(path.join(root,'.env'),'utf8');
const env=(key,fallback)=>process.env[key]??config.match(new RegExp(`^${key}=(.*)$`,'m'))?.[1].trim()??fallback;
const origin=(key,port)=>`http://127.0.0.1:${env(key,port)}`;
const web=origin('ORBIS_WEB_PORT','18080');
const jaeger=origin('ORBIS_JAEGER_PORT','18686');
const prometheus=origin('ORBIS_PROMETHEUS_PORT','18909');
const grafana=origin('ORBIS_GRAFANA_PORT','18304');
async function json(url,headers={}) {const res=await fetch(url,{headers,signal:AbortSignal.timeout(5000)});assert.equal(res.status,200,'Observability HTTP response');return res.json();}
async function eventually(action) {
  for(let i=0;i<30;i++) {try {await action();return;}catch {if(i===29) throw new Error('Observability did not converge within 60 seconds');await new Promise(r=>setTimeout(r,2000));}}
}
try {
  const response=await fetch(`${web}/api/v1/system`);assert.equal(response.status,200);
  const trace=response.headers.get('x-trace-id');assert.match(trace??'',/^[a-f0-9]{32}$/);
  await eventually(async()=>{
    const services=await json(`${jaeger}/api/services`);
    assert.ok(services.data.includes('orbis-api'));assert.ok(services.data.includes('orbis-worker'));
    const stored=await json(`${jaeger}/api/traces/${trace}`);assert.ok(stored.data.some(item=>item.traceID===trace));
  });
  await eventually(async()=>{
    const targets=await json(`${prometheus}/api/v1/targets`);
    for(const job of ['orbis-api','orbis-worker','prometheus']) assert.ok(targets.data.activeTargets.some(t=>t.labels.job===job&&t.health==='up'));
  });
  const password=env('ORBIS_GRAFANA_PASSWORD','');assert.ok(password);
  const headers={Authorization:`Basic ${Buffer.from(`admin:${password}`).toString('base64')}`};
  const dashboard=await json(`${grafana}/api/dashboards/uid/orbis-phase0`,headers);assert.equal(dashboard.dashboard.panels.length,4);
  const datasource=await json(`${grafana}/api/datasources/uid/orbis-prometheus`,headers);assert.equal(datasource.url,'http://prometheus:9090');
  fs.mkdirSync(path.join(root,'docs/reports'),{recursive:true});
  fs.writeFileSync(path.join(root,'docs/reports/phase0-observability.json'),JSON.stringify({date:new Date().toISOString(),status:'passed',checks:['API trace persisted in Jaeger','API/Worker service names','three Prometheus targets up','Grafana provisioned datasource and four panels']},null,2));
  console.log('PASS Jaeger persisted trace, Prometheus targets and Grafana provisioning');
} catch {console.error('Observability acceptance failed; ensure obs and --dev are enabled (credentials withheld)');process.exitCode=1;}
