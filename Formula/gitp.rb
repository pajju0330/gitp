class Gitp < Formula
  desc "Git CLI wrapper with named config profiles"
  homepage "https://github.com/pajju0330/gitp"
  url "https://github.com/pajju0330/gitp/archive/refs/tags/v0.1.0.tar.gz"
  # Update sha256 when cutting a release:
  #   curl -sL <url> | shasum -a 256
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"
  head "https://github.com/pajju0330/gitp.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/prajwal/gitp/internal/app.Version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/gitp"
  end

  test do
    assert_match "gitp", shell_output("#{bin}/gitp --version")
    assert_match "Usage", shell_output("#{bin}/gitp --help")
  end
end
