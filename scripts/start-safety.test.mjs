import assert from 'node:assert/strict';
import net from 'node:net';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('actual Compose preflight rejects an occupied Orbis entry without changing any container', async () => {
  const format='{{.ID}}|{{.Names}}|{{.State}}|{{.Ports}}';
  const before=execFileSync('docker',['ps','-a','--format',format],{encoding:'utf8'}).trim().split('\n').sort();
  const listener=net.createServer();
  await new Promise((resolve)=>listener.listen(0,'127.0.0.1',resolve));
  try {
    const result=await new Promise((resolve)=>{
      import('node:child_process').then(({execFile})=>execFile(process.execPath,['scripts/orbis.mjs','start','--profiles=app,rag','--no-build'],{
        env:{...process.env,ORBIS_WEB_PORT:String(listener.address().port)},
        timeout:30000,
      },(error,stdout,stderr)=>resolve({error,stdout,stderr})));
    });
    assert.equal(result.error?.code,1);
    assert.match(result.stderr,/proxy.*occupied/);
    assert.equal(listener.listening,true);
    const after=execFileSync('docker',['ps','-a','--format',format],{encoding:'utf8'}).trim().split('\n').sort();
    assert.deepEqual(after,before);
  } finally { await new Promise((resolve)=>listener.close(resolve)); }
});
