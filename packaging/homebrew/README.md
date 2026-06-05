# Homebrew Packaging

[日本語版](README.ja.md)

The formula template is intentionally not published yet. Release artifacts must exist before a tap can be updated.

Manual flow:

1. Create a GitHub release with `mockport_<version>_<os>_<arch>.tar.gz` archives and `checksums.txt`.
2. Replace `__VERSION__`, `__URL__`, and `__SHA256__` in `mockport.rb.template`.
3. Open a tap PR after verifying `brew install --build-from-source ./mockport.rb`.
