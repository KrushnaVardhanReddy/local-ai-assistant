#!/bin/bash
sed -i '/go system.StartBackgroundDownload/i \
	if a.remoteServer != nil {\
		a.remoteServer.Start()\
	}' wails-app/app.go

# add shutdown
cat << 'INNER_EOF' >> wails-app/app.go

func (a *App) shutdown(ctx context.Context) {
	if a.remoteServer != nil {
		_ = a.remoteServer.Stop(ctx)
	}
}
INNER_EOF
