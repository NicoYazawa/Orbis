import net from 'node:net';
import dgram from 'node:dgram';

export function validatePortMappings(mappings) {
  const seen = new Set();
  for (const mapping of mappings) {
    if (mapping.host !== '127.0.0.1') throw new Error(`${mapping.service}: host binding must be loopback 127.0.0.1`);
    if (!Number.isInteger(mapping.port) || mapping.port < 1 || mapping.port > 65535) throw new Error(`${mapping.service}: invalid host port`);
    if (!['tcp', 'udp'].includes(mapping.protocol)) throw new Error(`${mapping.service}: unsupported protocol`);
    const key = `${mapping.host}:${mapping.port}/${mapping.protocol}`;
    if (seen.has(key)) throw new Error(`${mapping.service}: duplicate host port ${key}`);
    seen.add(key);
  }
}

export async function assertPortsAvailable(mappings, owned = new Set()) {
  validatePortMappings(mappings);
  for (const mapping of mappings) {
    const key = `${mapping.service}:${mapping.port}/${mapping.protocol}`;
    if (owned.has(key)) continue;
    await new Promise((resolve, reject) => {
      const listener = mapping.protocol === 'udp' ? dgram.createSocket('udp4') : net.createServer();
      listener.once('error', (error) => reject(new Error(`${mapping.service}: port ${mapping.port}/${mapping.protocol} occupied or unavailable (${error.code}); other projects were not changed`)));
      const close = () => listener.close(resolve);
      if (mapping.protocol === 'udp') listener.bind(mapping.port, mapping.host, close);
      else listener.listen({port: mapping.port, host: mapping.host, exclusive: true}, close);
    });
  }
}
