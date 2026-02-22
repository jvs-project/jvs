# JVS Release Signing

All JVS releases are cryptographically signed using [Sigstore/cosign](https://github.com/sigstore/cosign) to provide authenticity and integrity verification.

## What is Signed?

Each release includes:

1. **Binaries** - Pre-built executables for multiple platforms
2. **Signatures** - `.sig` files containing digital signatures for each binary
3. **Certificates** - `.pem` files containing X.509 certificates from the signing workflow
4. **Checksums** - `SHA256SUMS` file containing SHA256 hashes of all binaries
5. **Checksums signature** - `SHA256SUMS.sig` and `SHA256SUMS.pem` for the checksums file itself

## Verification

### Installing cosign

Install the `cosign` tool:

```bash
# macOS/Linux (AMD64)
curl -O -L "https://github.com/sigstore/cosign/releases/latest/download/cosign-$(uname -s)-$(uname -m)"
chmod +x cosign
sudo mv cosign /usr/local/bin/

# Using go install
go install github.com/sigstore/cosign/v2/cmd/cosign@latest
```

### Verifying a Binary

To verify a downloaded binary:

```bash
# Download the binary, signature, and certificate
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/jvs-linux-amd64
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/jvs-linux-amd64.sig
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/jvs-linux-amd64.pem

# Verify using cosign
cosign verify-blob jvs-linux-amd64 \
  --signature jvs-linux-amd64.sig \
  --certificate jvs-linux-amd64.pem \
  --certificate-identity=https://github.com/jvs-project/jvs/.github/workflows/ci.yml@refs/tags/vX.Y.Z \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

Successful verification output:
```
Verified OK
```

### Verifying Checksums

To verify the SHA256SUMS file:

```bash
# Download checksums and signature
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/SHA256SUMS
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/SHA256SUMS.sig
wget https://github.com/jvs-project/jvs/releases/download/vX.Y.Z/SHA256SUMS.pem

# Verify the checksums file
cosign verify-blob SHA256SUMS \
  --signature SHA256SUMS.sig \
  --certificate SHA256SUMS.pem \
  --certificate-identity=https://github.com/jvs-project/jvs/.github/workflows/ci.yml@refs/tags/vX.Y.Z \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

Then verify your binary against the checksums:

```bash
sha256sum -c --ignore-missing SHA256SUMS
```

## Certificate Identity

All JVS releases are signed with the following certificate identity:

- **Identity**: `https://github.com/jvs-project/jvs/.github/workflows/ci.yml@refs/tags/vX.Y.Z`
- **Issuer**: `https://token.actions.githubusercontent.com`

This ensures the binary was built and signed by the official JVS CI workflow running on GitHub Actions.

## Manual Verification (without cosign)

If you prefer manual verification using GPG:

1. Download the `SHA256SUMS` file
2. Download your binary
3. Compare the SHA256 hash:

```bash
sha256sum -c SHA256SUMS
```

## Security Considerations

- **Keyless signing**: JVS uses Sigstore's keyless signing, which eliminates the need for managing private keys
- **OIDC identity**: Signatures are bound to GitHub Actions OIDC identity
- **Reproducibility**: All builds are performed in a transparent CI environment
- **Certificate transparency**: All signing events are recorded in the public Rekor transparency log

## Reporting Issues

If you encounter any verification issues or suspect a compromised release:

1. Do not run the binary
2. Report the issue immediately at [https://github.com/jvs-project/jvs/security/advisories](https://github.com/jvs-project/jvs/security/advisories)
3. Include the version, checksum, and any error messages
