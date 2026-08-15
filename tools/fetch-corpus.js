// Builds the corpus that tools/encoding-sizes.js measures.
//
// Usage: node tools/fetch-corpus.js [DIR]   (default: corpus/)
//
// The corpus lives in its own repository, attached here as a submodule, so a
// checkout of this repo carries a pinned reference and not 12 MB of samples.
// This tool is what produced that repository's contents, and re-running it
// reproduces them.
//
// Sources:
//   prose   Wikipedia via the Hugging Face datasets rows API (CC BY-SA 3.0)
//   code    Rosetta Code via the same API (GFDL 1.2)
//   images  the machine's own packaged files, with the package and the path
//           of its copyright file recorded in the category's CREDITS.json
//
// see docs/encodings.md

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');

const ROOT = process.argv[2] ?? 'corpus';
const PER_CATEGORY = Number(process.env.CORPUS_SAMPLES ?? 25);
const MAX_SAMPLE_BYTES = 20000;

// Wikipedia in scripts that stress the two modes differently: Latin-1 prose
// is what byte mode is for, and a script outside Latin-1 is what it is not.
const WIKIPEDIA = {
	'prose-english': '20231101.en',
	'prose-french': '20231101.fr',
	'prose-german': '20231101.de',
	'prose-vietnamese': '20231101.vi',
	'prose-russian': '20231101.ru',
	'prose-greek': '20231101.el',
	'prose-arabic': '20231101.ar',
	'prose-hindi': '20231101.hi',
	'prose-japanese': '20231101.ja',
	'prose-chinese': '20231101.zh',
};

const ROSETTA_LANGUAGES = ['Python', 'C', 'Java', 'JavaScript', 'Go', 'Haskell', 'Ruby', 'Rust'];

const SERVER = 'https://datasets-server.huggingface.co';

// A fixed retry cadence, and a failure that says which URL gave up rather
// than a corpus quietly missing a category.
// Wikimedia asks every client to identify itself and rate-limits the ones
// that do not, so the agent string is not decoration.
const AGENT = 'xml-validator-corpus/1.0 (https://github.com/wow-look-at-my/xml-validator)';

function curl(url) {
	let last;
	for (let attempt = 0; attempt < 4; attempt++) {
		try {
			return execFileSync('curl', ['-sSLf', '--max-time', '120', '-A', AGENT, url], { maxBuffer: 1 << 28 });
		} catch (e) {
			last = e;
			execFileSync('sleep', ['2']);
		}
	}
	throw new Error(`fetch failed after 4 attempts: ${url}\n${last?.stderr}`);
}

function write(category, name, data) {
	const dir = path.join(ROOT, category);
	fs.mkdirSync(dir, { recursive: true });
	fs.writeFileSync(path.join(dir, name), data);
}

function saveText(category, values, extension) {
	let n = 0;
	for (const value of values) {
		if (typeof value !== 'string' || value.length < 200) continue;
		if (n >= PER_CATEGORY) break;
		// Cap a sample so one long article cannot dominate its category.
		write(category, `${String(n).padStart(3, '0')}${extension}`, Buffer.from(value.slice(0, MAX_SAMPLE_BYTES), 'utf8'));
		n++;
	}
	console.log(`${category.padEnd(18)} ${n} samples`);
}

for (const [category, config] of Object.entries(WIKIPEDIA)) {
	const url = `${SERVER}/rows?dataset=wikimedia%2Fwikipedia&config=${config}&split=train&offset=0&length=${PER_CATEGORY}`;
	saveText(category, JSON.parse(curl(url)).rows.map((r) => r.row.text), '.txt');
}

for (const language of ROSETTA_LANGUAGES) {
	const where = encodeURIComponent(`"language_name"='${language}'`);
	const url = `${SERVER}/filter?dataset=christopher%2Frosetta-code&config=default&split=train&where=${where}&limit=${PER_CATEGORY}`;
	saveText(`code-${language.toLowerCase()}`, JSON.parse(curl(url)).rows.map((r) => r.row.code), '.txt');
}

// Images come from the distribution's own packages: real encoder output,
// already on the machine, and each one's package and license file is recorded
// beside it so the corpus can say where every sample came from.
function saveImages(extensions, category, count) {
	let listing = '';
	try {
		// Skip /tmp and the corpus itself: a sample copied from an earlier run
		// would make the corpus a picture of its own last state.
		const names = extensions.flatMap((e, i) => (i === 0 ? ['-name', `*.${e}`] : ['-o', '-name', `*.${e}`]));
		listing = execFileSync('find', ['/', '-xdev', '-path', '/tmp', '-prune', '-o',
			'-path', path.resolve(ROOT), '-prune', '-o',
			'(', ...names, ')', '-size', '+1k', '-size', '-2M', '-print'],
		{ encoding: 'utf8', maxBuffer: 1 << 28 });
	} catch (e) {
		listing = e.stdout ?? '';
	}
	const files = listing.split('\n').filter(Boolean);
	const step = Math.max(1, Math.floor(files.length / count));
	const picked = files.filter((_, i) => i % step === 0).slice(0, count);

	if (picked.length === 0) {
		throw new Error(`no ${extensions.join('/')} files found for ${category}`);
	}
	const credits = [];
	let n = 0;
	for (const file of picked) {
		let pkg = 'unknown';
		try {
			pkg = execFileSync('dpkg', ['-S', file], { encoding: 'utf8' }).split(':')[0];
		} catch {
			// A file no package claims still measures fine; the record says so.
		}
		const name = `${String(n).padStart(3, '0')}${path.extname(file)}`;
		write(category, name, fs.readFileSync(file));
		credits.push({ file: `${category}/${name}`, source: file, package: pkg, license: `/usr/share/doc/${pkg}/copyright` });
		n++;
	}
	fs.writeFileSync(path.join(ROOT, category, 'CREDITS.json'), JSON.stringify(credits, null, '\t') + '\n');
	console.log(`${category.padEnd(18)} ${n} samples`);
}

saveImages(['jpg', 'jpeg'], 'image-jpeg', PER_CATEGORY);
saveImages(['png'], 'image-png', PER_CATEGORY);
