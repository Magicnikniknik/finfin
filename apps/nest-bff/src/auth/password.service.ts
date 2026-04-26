import { Injectable } from '@nestjs/common';
import { randomBytes, scrypt as nodeScrypt, timingSafeEqual } from 'crypto';
import { promisify } from 'util';

const scrypt = promisify(nodeScrypt);

@Injectable()
export class PasswordService {
  async hash(password: string): Promise<string> {
    const salt = randomBytes(16);
    const derived = (await scrypt(password, salt, 64)) as Buffer;
    return `scrypt$${salt.toString('base64')}$${derived.toString('base64')}`;
  }

  async verify(password: string, encodedHash: string): Promise<boolean> {
    const [scheme, saltB64, hashB64] = encodedHash.split('$');
    if (scheme !== 'scrypt' || !saltB64 || !hashB64) {
      return false;
    }

    const salt = Buffer.from(saltB64, 'base64');
    const expected = Buffer.from(hashB64, 'base64');
    const actual = (await scrypt(password, salt, expected.length)) as Buffer;

    if (actual.length !== expected.length) {
      return false;
    }

    return timingSafeEqual(actual, expected);
  }
}
