// Run after `start --profiles=app,rag,obs --dev`; no signed URL is printed.
import { execFileSync, spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const text=fs.readFileSync(path.join(root,'.env'),'utf8');
const value=(key,fallback)=>process.env[key]??text.match(new RegExp(`^${key}=(.*)$`,'m'))?.[1].trim()??fallback;
const port=value('ORBIS_S3_PORT','18900');
const web=value('ORBIS_WEB_PORT','18080');
const args=['compose','-p','orbis','--env-file',path.join(root,'.env'),'-f',path.join(root,'compose.yaml'),'-f',path.join(root,'compose.dev.yaml')];
function run(extra){return execFileSync('docker',[...args,'run','--rm','--no-deps',...extra],{cwd:root,encoding:'utf8',stdio:['ignore','pipe','pipe']});}
let object;
try {
  object=JSON.parse(run(['-e',`ORBIS_S3_PUBLIC_ENDPOINT=http://127.0.0.1:${port}`,'verify','/app/verify','--presign-only']));
  if(!object.presigned_url||!object.object_key) throw new Error('Missing browser verification object');
  // Pipe browser output: trace/error details may contain the private signature.
  const browserEnv={...process.env,ORBIS_BASE_URL:`http://127.0.0.1:${web}`,ORBIS_S3_PRESIGN_URL:object.presigned_url};
  const localBrowsers=path.join(root,'frontend/.playwright');
  if(!browserEnv.PLAYWRIGHT_BROWSERS_PATH&&fs.existsSync(localBrowsers)) browserEnv.PLAYWRIGHT_BROWSERS_PATH=localBrowsers;
  const result=spawnSync(process.execPath,[path.join(root,'frontend/node_modules/@playwright/test/cli.js'),'test'],{
    cwd:path.join(root,'frontend'),encoding:'utf8',stdio:['ignore','pipe','pipe'],
    env:browserEnv
  });
  if(result.status!==0) throw new Error('Browser acceptance failed; inspect local frontend/playwright-report (contains private signed URL)');
  console.log('PASS browser UI and real presigned S3 fetch (2 tests, no skips)');
} catch(error) {console.error(error.message.startsWith('Browser acceptance')?error.message:'Browser setup failed; ensure Orbis is running with --dev (raw signatures withheld)');process.exitCode=1;}
finally {
  if(object?.object_key) {
    try {run(['verify','/app/verify','--delete-key',object.object_key]);}
    catch {console.error('Verification object cleanup failed; retry cleanup before ending acceptance');process.exitCode=1;}
  }
}
