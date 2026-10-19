// Downloads the corpus files listed in tools/corpus-manifest.tsv into corpus/.

const crypto = require('crypto');
const fs = require('fs');
const path = require('path');
const { Readable } = require('stream');
const { pipeline } = require('stream/promises');

const ROOT = path.join(__dirname, '..');
const MANIFEST = path.join(__dirname, 'corpus-manifest.tsv');
const CONCURRENCY = 4;

function sha256File(file) {
	return new Promise((resolve, reject) => {
		const hash = crypto.createHash('sha256');
		fs.createReadStream(file).on('data', (c) => hash.update(c)).on('error', reject)
			.on('end', () => resolve(hash.digest('hex')));
	});
}

async function present(file, sha256) {
	if (!fs.existsSync(file)) return false;
	return (await sha256File(file)) === sha256;
}

async function download(url, file, sha256) {
	const response = await fetch(url);
	if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
	fs.mkdirSync(path.dirname(file), { recursive: true });
	const partial = `${file}.part`;
	await pipeline(Readable.fromWeb(response.body), fs.createWriteStream(partial));
	const actual = await sha256File(partial);
	if (actual !== sha256) {
		fs.rmSync(partial);
		throw new Error(`${url}: sha256 ${actual}, manifest says ${sha256}`);
	}
	fs.renameSync(partial, file);
}

const entries = fs.readFileSync(MANIFEST, 'utf8').split('\n').filter(Boolean).map((line) => {
	const [sha256, dest, url] = line.split('\t');
	if (!sha256 || !dest || !url) throw new Error(`malformed manifest line: ${line}`);
	return { sha256, file: path.join(ROOT, dest), dest, url };
});

const only = process.argv.slice(2);
const wanted = only.length ? entries.filter((e) => only.some((prefix) => e.dest.startsWith(prefix))) : entries;
if (!wanted.length) throw new Error(`no manifest entries match ${only.join(' ')}`);

let next = 0;
let fetched = 0;
let skipped = 0;

async function worker() {
	while (next < wanted.length) {
		const entry = wanted[next++];
		if (await present(entry.file, entry.sha256)) {
			skipped++;
			continue;
		}
		await download(entry.url, entry.file, entry.sha256);
		fetched++;
		console.log(`fetched ${entry.dest}`);
	}
}

Promise.all(Array.from({ length: CONCURRENCY }, worker)).then(() => {
	console.log(`${fetched} fetched, ${skipped} already present`);
}, (err) => {
	console.error(err);
	process.exit(1);
});
