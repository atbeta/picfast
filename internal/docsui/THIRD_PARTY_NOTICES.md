# Third-Party Notices

## Scalar API Reference

- **Package:** `@scalar/api-reference`
- **Version:** 1.60.0
- **Source:** https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.60.0
- **License:** MIT
- **Website:** https://scalar.com

### Update Instructions

To update the Scalar bundle:

1. Download the new version:
   ```bash
   curl -sL "https://cdn.jsdelivr.net/npm/@scalar/api-reference@<VERSION>" -o internal/docsui/static/scalar-api-reference.js
   ```
2. Download fonts (if updated):
   ```bash
   for f in inter-cyrillic-ext inter-cyrillic inter-greek-ext inter-greek inter-latin-ext inter-latin inter-symbols inter-vietnamese mono-cyrillic-ext mono-cyrillic mono-greek-ext mono-greek mono-latin-ext mono-latin mono-vietnamese; do
     curl -sL "https://fonts.scalar.com/${f}.woff2" -o "internal/docsui/static/fonts/${f}.woff2"
   done
   ```
3. Replace font URLs in the bundle:
   ```bash
   sed -i '' 's|https://fonts\.scalar\.com/|/docs/assets/fonts/|g' internal/docsui/static/scalar-api-reference.js
   ```
4. Update the version in this file.
5. Run `make generate` if applicable and verify tests pass.

## Scalar Fonts

- **Font:** Inter (multiple subsets)
- **Source:** https://fonts.scalar.com
- **License:** SIL Open Font License 1.1
- **Files:** `static/fonts/*.woff2`
