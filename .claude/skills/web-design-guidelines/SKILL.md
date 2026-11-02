---
name: web-design-guidelines
description: Review UI code for Web Interface Guidelines compliance. Use when asked to "review my UI", "check accessibility", "audit design", "review UX", or "check my site against best practices".
metadata:
  author: vercel
  version: "1.0.0"
  argument-hint: <file-or-pattern>
---

# Web Interface Guidelines

Review files for compliance with Web Interface Guidelines.

## How It Works

1. Read the guidelines in [guidelines.md](guidelines.md), next to this file
2. Read the specified files (or prompt user for files/pattern)
3. Check against all rules in the guidelines
4. Output findings in the terse `file:line` format

## Precedence in SIGTA

`web/DESIGN.md` wins over these guidelines. In particular, interface text is in Rioplatense Spanish with sentence case ("Cerrar sesión"), so ignore the rules on Title Case and English copy, and styling goes through the Mantine theme, so a fix never adds CSS or Tailwind classes to a screen.

## Usage

When a user provides a file or pattern argument:
1. Read the guidelines
2. Read the specified files
3. Apply all rules from the guidelines
4. Output findings using the format specified in the guidelines

If no files specified, ask the user which files to review.
