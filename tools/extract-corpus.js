// Extracts measurable samples from the corpus submodules.
//
// Usage: node tools/extract-corpus.js [DIR]   (default: corpus/samples)
//
// The corpus is three Hugging Face datasets attached as submodules under
// corpus/. Each is a shallow, sparse, LFS-pointer-only clone, so a checkout of
// this repo costs pointers until someone fetches the blobs they want:
//
//	git submodule update --init --depth 1
//	git -C corpus/wikipedia    lfs pull --include="20231101.is/*,20231101.am/*"
//	git -C corpus/rosetta-code lfs pull --include="data/*"
//	git -C corpus/images       lfs pull --include="transformers/model_doc/*"
//
// Prose and code arrive as parquet, which duckdb reads. Images are files
// already, so tools/encoding-sizes.js reads them from the submodule directly
// and this tool leaves them alone.
//
// see docs/encodings.md

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');

const ROOT = process.argv[2] ?? path.join('corpus', 'samples');
const PER_CATEGORY = Number(process.env.CORPUS_SAMPLES ?? 25);
const MAX_SAMPLE_BYTES = 20000;
const CORPUS = path.join(__dirname, '..', 'corpus');

// Wikipedia in scripts that stress the two modes differently: Latin-1 prose
// is what byte mode is for, and a script outside Latin-1 is what it is not.
// These are the small language shards, so fetching them is cheap.
const WIKIPEDIA = {
	'prose-icelandic': '20231101.is',
	'prose-nepali': '20231101.ne',
	'prose-yiddish': '20231101.yi',
	'prose-amharic': '20231101.am',
};

const ROSETTA_LANGUAGES = ['Python', 'C', 'Java', 'JavaScript', 'Go', 'Haskell', 'Ruby', 'Rust'];

function duckdb(sql) {
	const out = execFileSync('duckdb', ['-json', '-c', sql], { encoding: 'utf8', maxBuffer: 1 << 28 });
	return out.trim() ? JSON.parse(out) : [];
}

function write(category, name, data) {
	const dir = path.join(ROOT, category);
	fs.mkdirSync(dir, { recursive: true });
	fs.writeFileSync(path.join(dir, name), data);
}

function save(category, values) {
	let n = 0;
	for (const value of values) {
		if (typeof value !== 'string' || value.length < 200) continue;
		if (n >= PER_CATEGORY) break;
		// Cap a sample so one long article cannot dominate its category.
		write(category, `${String(n).padStart(3, '0')}.txt`, Buffer.from(value.slice(0, MAX_SAMPLE_BYTES), 'utf8'));
		n++;
	}
	console.log(`${category.padEnd(18)} ${n} samples`);
}

// A pointer-only checkout is not a failure to report as an empty category:
// say which lfs pull command fills it in.
function requireBlobs(glob, hint) {
	const files = execFileSync('sh', ['-c', `ls -l ${glob} 2>/dev/null | awk '$5 > 100000'`], { encoding: 'utf8' });
	if (!files.trim()) {
		throw new Error(`no fetched blobs matching ${glob}\nrun: ${hint}`);
	}
}

for (const [category, config] of Object.entries(WIKIPEDIA)) {
	const glob = path.join(CORPUS, 'wikipedia', config, '*.parquet');
	requireBlobs(glob, `git -C corpus/wikipedia lfs pull --include="${config}/*"`);
	const rows = duckdb(`SELECT text FROM read_parquet('${glob}') LIMIT ${PER_CATEGORY * 4}`);
	save(category, rows.map((r) => r.text));
}

const rosetta = path.join(CORPUS, 'rosetta-code', 'data', '*.parquet');
requireBlobs(rosetta, 'git -C corpus/rosetta-code lfs pull --include="data/*"');
for (const language of ROSETTA_LANGUAGES) {
	const rows = duckdb(
		`SELECT code FROM read_parquet('${rosetta}') WHERE language_name = '${language}' LIMIT ${PER_CATEGORY * 4}`);
	save(`code-${language.toLowerCase()}`, rows.map((r) => r.code));
}

const images = path.join(CORPUS, 'images', 'transformers', 'model_doc');
if (fs.existsSync(images)) {
	const fetched = fs.readdirSync(images).filter((f) => fs.statSync(path.join(images, f)).size >= 1024);
	console.log(`images             ${fetched.length} files, measured in place from the submodule`);
} else {
	console.log('images             not checked out (see the header of this file)');
}
