#!/usr/bin/env node

/**
 * Generate Recova-compatible JWT for API testing.
 * Claims follow internal/modules/auth/token.go (SessionClaims + RegisteredClaims).
 */
const crypto = require("crypto");

function parseArg(name) {
  const prefix = `${name}=`;
  for (let i = 2; i < process.argv.length; i++) {
    const arg = process.argv[i];
    if (arg === name && i + 1 < process.argv.length) {
      return process.argv[i + 1];
    }
    if (arg.startsWith(prefix)) {
      return arg.slice(prefix.length);
    }
  }
  return null;
}

function base64url(input) {
  return Buffer.from(input)
    .toString("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");
}

function ensure(value, name) {
  const trimmed = String(value || "").trim();
  if (!trimmed) {
    throw new Error(`${name} is required`);
  }
  return trimmed;
}

function main() {
  const secret = ensure(
    parseArg("--secret") || process.env.JWT_SECRET,
    "JWT secret (via JWT_SECRET environment variable or --secret flag)",
  );
  const appName = ensure(
    parseArg("--issuer") || process.env.APP_NAME || "recova-backend-v2",
    "Issuer",
  );
  const audience = ensure(
    parseArg("--audience") || process.env.APP_URL || "http://localhost:3001",
    "Audience",
  );
  const userID = ensure(
    parseArg("--user") ||
      parseArg("--subject") ||
      "10000000-0000-0000-0000-000000000001",
    "Subject (user ID)",
  );
  const tokenType = String(parseArg("--type") || "access").trim();
  if (tokenType !== "access" && tokenType !== "refresh") {
    throw new Error("Token type must be access or refresh");
  }

  const rawTtl = parseArg("--ttl");
  const ttlSeconds = Number(rawTtl !== null ? rawTtl : 900);
  if (!Number.isFinite(ttlSeconds) || ttlSeconds <= 0) {
    throw new Error("TTL must be a positive number of seconds");
  }

  const outputJson = !process.argv.includes("--raw");

  const now = Math.floor(Date.now() / 1000);
  const payload = {
    token_type: tokenType,
    iss: appName,
    aud: [audience],
    sub: userID,
    iat: now,
    exp: now + ttlSeconds,
    jti: crypto.randomUUID(),
  };

  const header = { alg: "HS256", typ: "JWT" };
  const encodedHeader = base64url(JSON.stringify(header));
  const encodedPayload = base64url(JSON.stringify(payload));
  const unsigned = `${encodedHeader}.${encodedPayload}`;

  const signature = crypto
    .createHmac("sha256", secret)
    .update(unsigned)
    .digest("base64")
    .replace(/=/g, "")
    .replace(/\+/g, "-")
    .replace(/\//g, "_");

  const token = `${unsigned}.${signature}`;

  if (outputJson) {
    process.stdout.write(`${JSON.stringify({ token, payload }, null, 2)}\n`);
    return;
  }

  process.stdout.write(`${token}\n`);
}

try {
  main();
} catch (err) {
  process.stderr.write(`Error: ${err.message}\n`);
  process.exit(1);
}
