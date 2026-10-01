// Run only against the dedicated Orbis Phase 0 stack. Never addresses another project.
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'../..');
const compose=['compose','-p','orbis','--env-file',path.join(root,'.env'),'-f',path.join(root,'compose.yaml'),'--profile','app','--profile','rag','--profile','obs'];
function docker(extra) {return execFileSync('docker',[...compose,...extra],{cwd:root,encoding:'utf8',stdio:['ignore','pipe','pipe'],maxBuffer:2*1024*1024});}
function query(sql) {return docker(['exec','-T','postgres','psql','-U','orbis','-d','orbis','-Atc',sql]).trim();}
const envText=fs.readFileSync(path.join(root,'.env'),'utf8');
const port=Number(process.env.ORBIS_WEB_PORT??envText.match(/^ORBIS_WEB_PORT=(\d+)/m)?.[1]??18080);
const base=`http://127.0.0.1:${port}`;
const report={date:new Date().toISOString(),checks:[]};
async function check(name, action) {
  const start=Date.now();
  await action();
  report.checks.push({name,status:'passed',durationMs:Date.now()-start});
  console.log(`PASS ${name}`);
}
try {
  await check('system, readiness, metrics through same-origin proxy',async()=>{
    const sys=await fetch(`${base}/api/v1/system`);assert.equal(sys.status,200);
    const body=await sys.json();assert.ok(body.service);assert.ok(body.version);
    assert.equal((await fetch(`${base}/readyz`)).status,200);
    const metrics=await fetch(`${base}/metrics`);assert.equal(metrics.status,200);assert.match(await metrics.text(),/orbis_http_requests_total/);
  });
  await check('three repeated migrations preserve one administrator and its password hash',async()=>{
    const before=query("SELECT count(*)||':'||min(md5(password_hash)) FROM users WHERE role='admin'");
    assert.match(before,/^1:/);
    for(let i=0;i<3;i++) docker(['run','--rm','--no-deps','initialize']);
    const after=query("SELECT count(*)||':'||min(md5(password_hash)) FROM users WHERE role='admin'");
    assert.equal(after,before);
    assert.equal(query('SELECT dirty FROM schema_migrations'),'f');
  });
  await check('dependency loss returns 503 without losing liveness; restart recovers',async()=>{
    docker(['stop','redis']);
    try {
      assert.equal((await fetch(`${base}/healthz`)).status,200);
      assert.equal((await fetch(`${base}/readyz`)).status,503);
    } finally {docker(['start','redis']);}
    let recovered=false;
    for(let i=0;i<30;i++) {
      if((await fetch(`${base}/readyz`)).status===200){recovered=true;break;}
      await new Promise((resolve)=>setTimeout(resolve,1000));
    }
    assert.equal(recovered,true);
  });
  await check('runtime logs do not contain generated credentials',async()=>{
    const log=docker(['logs','--no-color','api','worker','initialize']);
    for(const key of ['ORBIS_POSTGRES_PASSWORD','ORBIS_S3_ACCESS_KEY','ORBIS_S3_SECRET_KEY','ORBIS_ADMIN_PASSWORD']) {
      const value=envText.match(new RegExp(`^${key}=(.+)$`,'m'))?.[1].trim();
      assert.ok(value);assert.equal(log.includes(value),false,`${key} leaked into logs`);
    }
  });
} catch(error) {
  report.status='failed';report.failure=error instanceof assert.AssertionError?error.message:'acceptance command failed (raw credentials withheld)';
  console.error(report.failure);process.exitCode=1;
} finally {
  report.status??='passed';
  fs.mkdirSync(path.join(root,'docs/reports'),{recursive:true});
  fs.writeFileSync(path.join(root,'docs/reports/phase0-acceptance.json'),JSON.stringify(report,null,2));
}
