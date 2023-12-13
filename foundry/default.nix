{ lib
, stdenv
, fetchFromGitHub
, rustPlatform
, rustc
, libusb1
, darwin
, solc-bin-src
}:

rustPlatform.buildRustPackage rec {
  pname = "foundry";
  version = "0.2.0-dev";

  # To update run the following command and change `rev` and `hash`
  # nix-prefetch-github foundry-rs foundry --nix --rev 3e962e2efe17396886fcb1fd141ccf4204cd3a21
  src = fetchFromGitHub {
    owner = "foundry-rs";
    repo = "foundry";
    rev = "3e962e2efe17396886fcb1fd141ccf4204cd3a21";
    hash = "sha256-4ESOsNYtTPSnvtBtcXZ5FJXGCQOpW1mtlVOLyeJnLCE=";
  };

  cargoLock = {
    lockFile = src + "/Cargo.lock";
    allowBuiltinFetchGit = true;
  };

  nativeBuildInputs = [
    libusb1
  ] ++ lib.optionals stdenv.isDarwin [ darwin.DarwinTools ];

  buildInputs = lib.optionals stdenv.isDarwin [ darwin.apple_sdk.frameworks.AppKit ];

  # Tests fail
  doCheck = false;

  # Enable svm-rs to build without network access by giving it a list of solc releases.
  # We take the list from our solc flake repo so that it does not update unless we want to.
  # Our repo: https://github.com/EspressoSystems/nix-solc-bin
  # Upstream URL: https://binaries.soliditylang.org/linux-amd64/list.json
  env = let platform = if stdenv.isLinux then "linux-amd64" else "macosx-amd64"; in {
    SVM_RELEASES_LIST_JSON = "${solc-bin-src}/list-${platform}.json";
    # Make `vergen` produce a meaningful version.
    VERGEN_BUILD_TIMESTAMP = "0";
    VERGEN_BUILD_SEMVER = version;
    VERGEN_GIT_SHA = src.rev;
    VERGEN_GIT_COMMIT_TIMESTAMP = "0";
    VERGEN_GIT_BRANCH = "master";
    VERGEN_RUSTC_SEMVER = rustc.version;
    VERGEN_RUSTC_CHANNEL = "stable";
    VERGEN_CARGO_PROFILE = "release";
  };

  meta = with lib; {
    description = "A blazing fast, portable and modular toolkit for Ethereum application development";
    homepage = "https://github.com/foundry-rs/foundry";
    license = with licenses; [ mit apsl20 ];
    maintainers = [ ];
  };
}
