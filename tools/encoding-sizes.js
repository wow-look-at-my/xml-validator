// Measures what each XML wire form costs, over a corpus of real files.
//
// Usage: node tools/encoding-sizes.js [--validate] [--samples N]
//
// Every form carries the same payload: the file's bytes. A document is built,
// optionally checked with build/xml-validator, and its size recorded. The
// table it prints is the one in docs/encodings.md.
//
// see docs/encodings.md

const fs = require('fs');
const path = require('path');
const { execFileSync, spawnSync } = require('child_process');

const MAX_BYTES = 32 * 1024;
const PER_CATEGORY = Number(argValue('--samples', 40));
const VALIDATE = process.argv.includes('--validate');
const BINARY = path.resolve('build/xml-validator');

function argValue(flag, fallback) {
	const at = process.argv.indexOf(flag);
	return at >= 0 && process.argv[at + 1] ? process.argv[at + 1] : fallback;
}

// With --corpus DIR, every subdirectory of DIR is a category and every file
// in it is a sample -- the layout tools/fetch-corpus.js writes.
function corpusCategories(root) {
	const out = {};
	for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
		if (!entry.isDirectory()) continue;
		if (entry.name.startsWith('.')) continue;
		const files = fs.readdirSync(path.join(root, entry.name))
			.map((f) => path.join(root, entry.name, f))
			.filter((f) => fs.statSync(f).isFile() && !f.endsWith('CREDITS.json'));
		if (files.length > 0) out[entry.name] = files;
	}
	return out;
}

// Without one, each category is a find(1) expression over the machine's own
// files, so the corpus is still real data rather than something invented here.
const FIND_CATEGORIES = {
	'Go source': ['/home/user/xml-validator', '-name', '*.go'],
	'JavaScript/TS': ['/opt', '-name', '*.js', '-o', '-name', '*.ts'],
	'JSON': ['/opt', '-name', '*.json'],
	'Markdown': ['/root/.claude', '-name', '*.md'],
	'YAML': ['/opt', '-name', '*.yml', '-o', '-name', '*.yaml'],
	'plain text': ['/usr/share/doc', '-name', 'copyright'],
	'PNG image': ['/', '-xdev', '-name', '*.png'],
	'JPEG image': ['/', '-xdev', '-name', '*.jpg', '-o', '-name', '*.jpeg'],
	'SVG image': ['/', '-xdev', '-name', '*.svg'],
	'ELF binary': ['/usr/bin', '-type', 'f'],
	'gzip data': ['/usr/share', '-name', '*.gz'],
};

function collect(findArgs) {
	let out = '';
	try {
		out = execFileSync('find', [...findArgs, '-type', 'f'], { encoding: 'utf8', maxBuffer: 1 << 28 });
	} catch (e) {
		out = e.stdout ?? '';
	}
	const files = [];
	for (const p of out.split('\n')) {
		if (!p) continue;
		let st;
		try {
			st = fs.statSync(p);
		} catch {
			continue;
		}
		if (!st.isFile() || st.size === 0 || st.size > MAX_BYTES) continue;
		files.push(p);
		if (files.length >= PER_CATEGORY * 4) break;
	}
	// Spread the sample across the listing instead of taking the first N,
	// which would be one directory's worth of near-identical files.
	const step = Math.max(1, Math.floor(files.length / PER_CATEGORY));
	return files.filter((_, i) => i % step === 0).slice(0, PER_CATEGORY);
}

const XML_DECL = '<?xml version="1.1"?>';
const BYTE_DECL = '<?xml version="1.1" encoding="ISO-8859-1"?>';

function restricted(c) {
	return (c >= 0x01 && c <= 0x08) || (c >= 0x0b && c <= 0x0c) || (c >= 0x0e && c <= 0x1f) ||
		(c >= 0x7f && c <= 0x84) || (c >= 0x86 && c <= 0x9f);
}

// A character that cannot stand for itself: markup delimiters, U+0000, the
// restricted characters, and the three that line-ending normalization folds.
function needsReference(c) {
	return c === 0x26 || c === 0x3c || c === 0x3e || c === 0 ||
		restricted(c) || c === 0x0d || c === 0x85 || c === 0x2028;
}

function escapeChar(c) {
	if (c === 0x26) return '&amp;';
	if (c === 0x3c) return '&lt;';
	if (c === 0x3e) return '&gt;';
	return `&#${c};`;
}

function allReferences(payload) {
	const parts = [XML_DECL, '<r>'];
	for (const b of payload) parts.push(`&#${b};`);
	parts.push('</r>');
	return Buffer.from(parts.join(''), 'latin1');
}

function byteMode(payload) {
	const parts = [BYTE_DECL, '<r>'];
	for (const b of payload) parts.push(needsReference(b) ? escapeChar(b) : String.fromCharCode(b));
	parts.push('</r>');
	return Buffer.from(parts.join(''), 'latin1');
}

// UTF-8 only carries a payload that already IS valid UTF-8; anything else has
// no character to write. U+FFFE and U+FFFF are not characters at all, so a
// payload containing one has no UTF-8 form either.
function utf8Mode(payload) {
	const text = payload.toString('utf8');
	if (!Buffer.from(text, 'utf8').equals(payload)) return null;
	const parts = [XML_DECL, '<r>'];
	for (const ch of text) {
		const c = ch.codePointAt(0);
		if ((c & 0xfffe) === 0xfffe) return null;
		parts.push(needsReference(c) ? escapeChar(c) : ch);
	}
	parts.push('</r>');
	return Buffer.from(parts.join(''), 'utf8');
}

// Byte mode carrying TEXT rather than bytes: the payload is decoded, and each
// character it holds is written as one byte where Latin-1 has one, and as a
// reference where it does not. This is the mode's intended use, and the
// column above is the same mode asked to carry arbitrary bytes.
function byteModeText(payload) {
	const text = payload.toString('utf8');
	if (!Buffer.from(text, 'utf8').equals(payload)) return null;
	const parts = [BYTE_DECL, '<r>'];
	for (const ch of text) {
		const c = ch.codePointAt(0);
		if ((c & 0xfffe) === 0xfffe) return null;
		if (c > 0xff || needsReference(c)) parts.push(escapeChar(c));
		else parts.push(String.fromCharCode(c));
	}
	parts.push('</r>');
	return Buffer.from(parts.join(''), 'latin1');
}

function base64Mode(payload) {
	return Buffer.from(`${XML_DECL}<blob>${payload.toString('base64')}</blob>`, 'latin1');
}

function hexMode(payload) {
	return Buffer.from(`${XML_DECL}<blob>${payload.toString('hex').toUpperCase()}</blob>`, 'latin1');
}

const FORMS = {
	'byte mode (text)': byteModeText,
	'byte mode (bytes)': byteMode,
	'UTF-8': utf8Mode,
	'base64Binary': base64Mode,
	'hexBinary': hexMode,
	'all references': allReferences,
};

function validate(doc) {
	const res = spawnSync(BINARY, [], { input: doc });
	if (res.status !== 0) {
		throw new Error(`validator rejected a generated document: ${res.stderr}`);
	}
}

const totals = {};
const totalPayload = {};
const rows = [];
let sampleCount = 0;
let validated = 0;

const corpusDir = argValue('--corpus', null);
const categories = corpusDir
	? corpusCategories(corpusDir)
	: Object.fromEntries(Object.entries(FIND_CATEGORIES).map(([k, v]) => [k, collect(v)]));

for (const [category, all] of Object.entries(categories)) {
	const files = all.slice(0, PER_CATEGORY);
	if (files.length === 0) continue;
	const sums = { payload: 0 };
	const counts = {};
	for (const file of files) {
		const payload = fs.readFileSync(file);
		sampleCount++;
		sums.payload += payload.length;
		for (const [form, build] of Object.entries(FORMS)) {
			const doc = build(payload);
			if (doc === null) continue;
			if (VALIDATE) {
				validate(doc);
				validated++;
			}
			sums[form] = (sums[form] ?? 0) + doc.length;
			counts[form] = (counts[form] ?? 0) + payload.length;
			totals[form] = (totals[form] ?? 0) + doc.length;
			// Per form, because a form skips the samples it cannot carry and
			// dividing by the whole corpus would report it as free.
			totalPayload[form] = (totalPayload[form] ?? 0) + payload.length;
		}
		totals.payload = (totals.payload ?? 0) + payload.length;
	}
	rows.push({ category, files: files.length, sums, counts });
}

const forms = Object.keys(FORMS);
console.log(`# ${sampleCount} samples, ${(totals.payload / 1024).toFixed(0)} KiB of payload` +
	(VALIDATE ? `, ${validated} documents validated` : ' (sizes only, pass --validate to check each)'));
console.log();
console.log(`| corpus | files | median size | ${forms.join(' | ')} |`);
console.log(`|---|---:|---:|${forms.map(() => '---:').join('|')}|`);
for (const row of rows) {
	const cells = forms.map((f) => {
		if (!row.sums[f]) return 'n/a';
		return `${(row.sums[f] / row.counts[f]).toFixed(2)}x`;
	});
	const avg = Math.round(row.sums.payload / row.files);
	console.log(`| ${row.category} | ${row.files} | ${avg} B | ${cells.join(' | ')} |`);
}
const overall = forms.map((f) => `${(totals[f] / totalPayload[f]).toFixed(2)}x`);
console.log(`| **all** | **${sampleCount}** | | ${overall.map((c) => `**${c}**`).join(' | ')} |`);
console.log();
console.log('Ratios are document bytes per payload byte, so lower is smaller.');
console.log('"n/a" means no sample in that corpus had a form: UTF-8 cannot carry a');
console.log('payload that is not already valid UTF-8.');
