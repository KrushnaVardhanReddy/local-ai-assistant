import sys

content = open("wails-app/app_test.go").read()

if "TestSetProxyToken" not in content:
    content += """
func TestSetProxyToken(t *testing.T) {
	app := NewApp()
	app.SetProxyToken("test-token")

	if app == nil {
		t.Errorf("App is nil")
	}
}
"""

open("wails-app/app_test.go", "w").write(content)
