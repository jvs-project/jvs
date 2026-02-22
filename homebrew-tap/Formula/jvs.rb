# typed: false
# frozen_string_literal: true

class Jvs < Formula
  desc "Juicy Versioned Workspaces - Snapshot-based versioning for JuiceFS"
  homepage "https://github.com/jvs-project/jvs"
  url "https://github.com/jvs-project/jvs/archive/refs/tags/v7.0.tar.gz"
  sha256 "f1e8d0e0d8b7c9e5a6b3d7c8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8"
  license "MIT"

  bottle do
    root_url "https://github.com/jvs-project/jvs/releases/download/v7.0"
    sha256 cellar: :any_skip_relocation, arm64_sonoma:   "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, arm64_ventura:  "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, arm64_monterey: "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, sonoma:         "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, ventura:        "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, monterey:       "0000000000000000000000000000000000000000000000000000000000000000"
    sha256 cellar: :any_skip_relocation, x86_64_linux:   "0000000000000000000000000000000000000000000000000000000000000000"
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/jvs"

    # Install shell completion files
    generate_completions_from_executable(bin/"jvs", "completion", "bash")
    generate_completions_from_executable(bin/"jvs", "completion", "zsh")
    generate_completions_from_executable(bin/"jvs", "completion", "fish")
  end

  test do
    # Test basic functionality
    system bin/"jvs", "version"

    # Test help command
    system bin/"jvs", "--help"

    # Test that we can create a temporary repository
    testdir = testpath/"test_repo"
    testdir.mkpath
    cd testdir do
      system bin/"jvs", "init", "test"
      assert_predicate testdir/".jvs", :exist?
    end
  end
end
