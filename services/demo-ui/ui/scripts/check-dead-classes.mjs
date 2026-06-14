#!/usr/bin/env node
/**
 * Guard against Tailwind/Bootstrap utility classes creeping back into the
 * codebase. This project has NO Tailwind build step — such classes are inert
 * (they render no styles). Use the semantic design-system classes in
 * src/index.css instead.
 *
 * Runs over src/, fails (exit 1) if any banned utility class is found in a
 * className/class string. Run via `yarn lint:classes`.
 */
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, extname } from 'node:path';
import { fileURLToPath } from 'node:url';

const SRC = join(fileURLToPath(new URL('.', import.meta.url)), '..', 'src');

// Patterns that indicate inert Tailwind/Bootstrap utilities. Kept deliberately
// specific to avoid flagging legitimate semantic class names.
// NOTE: a small set of Bootstrap-ish spacing/text utilities (text-muted, py-4,
// d-block, mb-3, p-2 …) are deliberately shimmed in src/index.css and are NOT
// banned here. These patterns target only classes with no backing CSS rule.
const BANNED = [
    /\bbg-(blue|red|green|gray|slate|indigo|emerald|amber|white|black)-\d{2,3}\b/,
    /\btext-(blue|red|green|gray|slate|indigo|white|black)-\d{2,3}\b/,
    /\bhover:[a-z-]+/,
    /\brounded-(sm|md|lg|xl|2xl|full)\b/,
    /\bshadow-(sm|md|lg|xl|2xl)\b/,
    /\b(min-h-screen|min-w-screen)\b/,
    /\bgrid-cols-\d+\b/,
    /\b(justify|items)-(center|between|start|end|around)\b/,
    /\bspace-(x|y)-\d\b/,
    /\b(w|h)-(full|screen)\b/,
    /\bbg-gradient-to-[a-z]+\b/,
    /\b(d-flex|text-decoration-none|align-items-center)\b/, // bootstrap (un-shimmed)
];

// Only inspect class strings to avoid false positives in normal code/comments.
const CLASS_ATTR = /\bclassName\s*=\s*(?:"([^"]*)"|'([^']*)'|\{`([^`]*)`\})/g;

let violations = [];

function walk(dir) {
    for (const entry of readdirSync(dir)) {
        const full = join(dir, entry);
        if (statSync(full).isDirectory()) {
            walk(full);
        } else if (['.tsx', '.ts', '.jsx'].includes(extname(full))) {
            scan(full);
        }
    }
}

function scan(file) {
    // Skip tests and Storybook demos — neither ships in the app bundle.
    if (/\.(test|stories)\.(tsx?|jsx)$/.test(file)) return;
    const text = readFileSync(file, 'utf8');
    let m;
    while ((m = CLASS_ATTR.exec(text)) !== null) {
        const classes = m[1] ?? m[2] ?? m[3] ?? '';
        for (const pattern of BANNED) {
            const hit = classes.match(pattern);
            if (hit) {
                const line = text.slice(0, m.index).split('\n').length;
                violations.push(`${file}:${line}  banned utility class "${hit[0]}" in: ${classes.trim()}`);
            }
        }
    }
}

walk(SRC);

if (violations.length > 0) {
    console.error('✖ Inert Tailwind/Bootstrap utility classes found (no Tailwind build exists):\n');
    console.error(violations.join('\n'));
    console.error(`\n${violations.length} violation(s). Use semantic classes from src/index.css instead.`);
    process.exit(1);
}

console.log('✓ No banned utility classes found.');
