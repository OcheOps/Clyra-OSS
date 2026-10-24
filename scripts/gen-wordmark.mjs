#!/usr/bin/env node
// Regenerate the clyra wordmark PNGs at three background variants
// (transparent / slate / white) × four heights (256/512/1024/2048)
// and write them to web/public/brand/.
//
// Font is Switzer 600 from Fontshare (matches the in-app wordmark
// exactly), rendered through headless Chrome. Requires puppeteer-core
// and a system Chrome/Chromium at /usr/bin/google-chrome.
//
// Run:  node scripts/gen-wordmark.mjs
//
import puppeteer from 'puppeteer-core';
import { writeFile } from 'node:fs/promises';

const sizes = [256, 512, 1024, 2048];
const variants = [
  { suffix: '',        bg: 'transparent' },
  { suffix: '-slate',  bg: '#0F172A'     },
  { suffix: '-white',  bg: '#FFFFFF', fg: '#0F172A' },
];

const browser = await puppeteer.launch({
  executablePath: '/usr/bin/google-chrome',
  args: ['--no-sandbox', '--headless=new'],
});
const page = await browser.newPage();

for (const v of variants) {
  for (const h of sizes) {
    const fontSize = Math.round(h * 0.55);
    const padY = Math.round(h * 0.18);
    const padX = Math.round(h * 0.24);
    const fg = v.fg || '#F8FAFC';
    const html = `<!doctype html><html><head>
<link rel="stylesheet" href="https://api.fontshare.com/v2/css?f[]=switzer@600&display=swap">
<style>
  html,body { margin:0; padding:0; }
  html { background: ${v.bg}; }
  .box { display:inline-block; padding:${padY}px ${padX}px; background:${v.bg};
         font-family:'Switzer',system-ui,sans-serif; font-weight:600; letter-spacing:-0.015em;
         font-size:${fontSize}px; line-height:1; color:${fg}; }
  .dot { color:#06B6D4; }
</style></head><body><div class="box">clyra<span class="dot">.</span></div></body></html>`;
    await page.setViewport({ width: 4000, height: h + 200, deviceScaleFactor: 1 });
    await page.setContent(html, { waitUntil: 'domcontentloaded' });
    await page.evaluate(async () => { await document.fonts.ready; });
    const el = await page.$('.box');
    const buf = await el.screenshot({ omitBackground: v.suffix === '' });
    const name = `clyra-wordmark-${h}${v.suffix}.png`;
    await writeFile('./web/public/brand/' + name, buf);
    console.log(name, buf.length, 'bytes');
  }
}
await browser.close();
