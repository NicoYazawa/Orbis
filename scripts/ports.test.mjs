import assert from 'node:assert/strict';
import net from 'node:net';
import { test } from 'node:test';
import { assertPortsAvailable, validatePortMappings } from './ports.mjs';

test('occupied host port fails without touching its listener', async () => {
  const listener = net.createServer();
  await new Promise((resolve) => listener.listen(0, '127.0.0.1', resolve));
  try {
    const port = listener.address().port;
    await assert.rejects(assertPortsAvailable([{ service: 'proxy', host: '127.0.0.1', port, protocol: 'tcp' }]), /proxy.*occupied/);
    assert.equal(listener.listening, true);
  } finally { await new Promise((resolve) => listener.close(resolve)); }
});

test('ports assigned to two Orbis services fail before Docker starts', () => {
  assert.throws(() => validatePortMappings([
    { service: 'proxy', host: '127.0.0.1', port: 18080, protocol: 'tcp' },
    { service: 'api', host: '127.0.0.1', port: 18080, protocol: 'tcp' },
  ]), /duplicate/);
});

test('unsafe host binding and invalid ports are rejected', () => {
  assert.throws(() => validatePortMappings([{service:'api',host:'0.0.0.0',port:18081,protocol:'tcp'}]), /loopback/);
  assert.throws(() => validatePortMappings([{service:'api',host:'127.0.0.1',port:0,protocol:'tcp'}]), /invalid/);
});
