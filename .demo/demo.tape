# n8nctl demo — rendered with VHS (https://github.com/charmbracelet/vhs).
# Records the locally-built binary against a local mock API for clean,
# reproducible output (no real instance or secrets). Reproduce with:
#
#   make build
#   export PATH=$PWD/bin:$PATH
#   export N8NCTL_CONFIG=$PWD/.demo/config.yaml
#   (python3 .demo/mockserver.py >/dev/null 2>&1 &) ; sleep 1
#   vhs .demo/demo.tape
#   pkill -f mockserver.py

Output assets/demo.gif

Set Shell bash
Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 18
Set Theme "Dracula"
Set TypingSpeed 25ms

Hide
Ctrl+L
Show

Sleep 800ms
Type "n8nctl workflows list --quiet"
Sleep 400ms
Enter
Sleep 1800ms

Type "# lint workflow files before they ship — grounded in n8n's own schema"
Sleep 250ms
Enter
Type "n8nctl workflows lint -f .demo/needs-review.json"
Sleep 400ms
Enter
Sleep 2600ms

Type "# GitOps: a directory of files is the desired state — preview with --dry-run"
Sleep 250ms
Enter
Type "n8nctl workflows apply --dir .demo/workflows --dry-run"
Sleep 400ms
Enter
Sleep 2200ms

Type "# search every workflow's node graph (impossible in the UI)"
Sleep 250ms
Enter
Type "n8nctl workflows search --node httpRequest --quiet"
Sleep 400ms
Enter
Sleep 2200ms

Type "# snapshot the instance to git-friendly YAML"
Sleep 250ms
Enter
Type "n8nctl backup --out /tmp/n8n-backup --format yaml --quiet && ls /tmp/n8n-backup/workflows"
Sleep 400ms
Enter
Sleep 2600ms
