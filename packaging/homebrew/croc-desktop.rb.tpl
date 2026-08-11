# Homebrew cask template for croc-desktop.
# release.yml (publish-homebrew job) renders {{VERSION}} and {{SHA256}} from the
# tag and the freshly built DMG, then pushes the result to
# SihanTeng/homebrew-tap. Do not hand-edit versions here.
cask "croc-desktop" do
  version "{{VERSION}}"
  sha256 "{{SHA256}}"

  url "https://github.com/SihanTeng/croc-desktop/releases/download/v#{version}/croc-desktop_v#{version}_darwin-arm64.dmg"
  name "croc-desktop"
  desc "Send files and text between devices — GUI for croc"
  homepage "https://github.com/SihanTeng/croc-desktop"

  depends_on arch: :arm64
  depends_on macos: :monterey

  app "croc-desktop.app"

  zap trash: [
    "~/.config/croc/croc-desktop.json",
    "~/.config/croc/croc-desktop-history.json",
    "~/.config/croc/croc-desktop.log",
  ]
end
