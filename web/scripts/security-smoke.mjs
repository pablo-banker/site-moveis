import assert from 'node:assert/strict';
const base = process.env.SECURITY_TEST_URL || 'http://127.0.0.1:5173';
async function request(path, options, expected, allowBodyAbort = false) {
	let response;
	try {
		response = await fetch(base + path, { ...options, redirect: 'manual' });
	} catch (error) {
		// Cancelling a chunked body destroys the Node adapter socket; it must never reach the API.
		if (allowBodyAbort && ['ECONNRESET', 'UND_ERR_SOCKET'].includes(error.cause?.code)) return null;
		throw error;
	}
	assert.equal(
		response.status,
		expected,
		`${path}: status ${response.status}, expected ${expected}`,
	);
	await response.text();
	return response;
}
const first = await request('/', {}, 200);
assert.equal(first.headers.get('x-content-type-options'), 'nosniff');
assert.equal(first.headers.get('x-frame-options'), 'DENY');
const csp = first.headers.get('content-security-policy');
assert(csp?.includes("script-src 'self'"));
assert(csp?.includes("frame-ancestors 'none'"));
assert(!csp?.match(/script-src[^;]*unsafe-inline/));
const second = await request('/', {}, 200);
assert.notEqual(csp, second.headers.get('content-security-policy'), 'SSR CSP nonce must change');
await request(
	'/backend/auth/logout',
	{
		method: 'POST',
		headers: { Origin: 'https://attacker.invalid', 'Content-Type': 'application/json' },
		body: '{}',
	},
	403,
);
await request(
	'/backend/auth/logout',
	{
		method: 'POST',
		headers: { Origin: base, 'Sec-Fetch-Site': 'cross-site', 'Content-Type': 'application/json' },
		body: '{}',
	},
	403,
);
await request(
	'/backend/auth/logout',
	{ method: 'POST', headers: { Origin: base, 'Content-Type': 'text/plain' }, body: '{}' },
	415,
);
await request(
	'/backend/auth/login',
	{
		method: 'POST',
		headers: { Origin: base, 'Content-Type': 'application/json' },
		body: JSON.stringify({ email: 'x'.repeat(40000) }),
	},
	413,
);
// A chunked body must be counted by bytes, not JavaScript characters.
const body = new ReadableStream({
	start(c) {
		c.enqueue(new TextEncoder().encode(JSON.stringify({ email: 'é'.repeat(20000) })));
		c.close();
	},
});
await request(
	'/backend/auth/login',
	{
		method: 'POST',
		headers: { Origin: base, 'Content-Type': 'application/json' },
		body,
		duplex: 'half',
	},
	413,
	true,
);
await request('/backend/admin', {}, 404);
const invalid = await request(
	'/backend/auth/me',
	{ headers: { Cookie: 'forma-session=invalid' } },
	401,
);
assert.match(invalid.headers.get('cache-control'), /no-store/);
const login = await request(
	'/backend/auth/login',
	{
		method: 'POST',
		headers: { Origin: base, 'Content-Type': 'application/json', Cookie: 'forma-session=invalid' },
		body: JSON.stringify({
			email: 'security-nonexistent@example.invalid',
			password: 'WrongPassword123456!',
		}),
	},
	401,
);
assert.equal(
	login.headers.get('set-cookie'),
	null,
	'Failed login must not clear an existing session',
);
console.log(
	'Security HTTP smoke: headers/CSP, CSRF, JSON-only, byte/chunk limits, route allowlist, session cache and login isolation passed.',
);
