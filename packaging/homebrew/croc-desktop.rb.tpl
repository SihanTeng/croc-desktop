# Homebrew cask template for croc-desktop.
# release.yml (publish-homebrew job) renders the version and sha256 from the
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

  # Releases are ad-hoc signed (no Developer ID/notarization yet), and
  # Gatekeeper refuses quarantined ad-hoc apps. Strip the quarantine
  # attribute at install so the app opens on first try (same approach as
  # kage's cask). Remove this once releases are notarized.
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-dr", "com.apple.quarantine", "#{appdir}/croc-desktop.app"]
  end

  zap trash: [
    "~/.config/croc/croc-desktop.json",
    "~/.config/croc/croc-desktop-history.json",
    "~/.config/croc/croc-desktop.log",
  ]
end
