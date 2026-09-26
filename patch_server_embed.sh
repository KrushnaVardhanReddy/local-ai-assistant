cat << 'INNER_EOF' >> wails-app/backend/buddy/server.go

//go:embed index.html
var staticAssets embed.FS
INNER_EOF
