import sharp from 'sharp';
import pngToIco from 'png-to-ico';
import { readFile, writeFile, mkdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const SRC_SVG = path.resolve(__dirname, '../src/assets/brand/logo-primary.svg');
const OUT_DIR  = path.resolve(__dirname, '../public');

async function ensureOutDir() {
	await mkdir(OUT_DIR, { recursive: true });
}

const out = (name) => path.join(OUT_DIR, name);

(async () => {
	await ensureOutDir();

	const svg = await readFile(SRC_SVG);

	// PNGs for PWA / iOS
	await sharp(svg).resize(180, 180).png().toFile(out('apple-touch-icon.png'));
	await sharp(svg).resize(192, 192).png().toFile(out('icon-192.png'));
	await sharp(svg).resize(512, 512).png().toFile(out('icon-512.png'));

	// Favicon ICO: build from multiple PNG sizes for best results
	const sizes = [16, 24, 32, 48];
	const pngBuffers = await Promise.all(
		sizes.map((s) => sharp(svg).resize(s, s).png().toBuffer())
	);
	const icoBuffer = await pngToIco(pngBuffers);
	await writeFile(out('favicon.ico'), icoBuffer);

	// Also ship an SVG favicon for modern browsers
	await writeFile(out('favicon.svg'), svg);
})();