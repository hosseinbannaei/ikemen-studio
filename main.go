package main

/*
#cgo linux CFLAGS: -D_GNU_SOURCE
#include <stdlib.h>
#include <string.h>

__attribute__((constructor))
static void fix_linux_environment(void) {
    // 1. Sanitize LD_LIBRARY_PATH from Snap-injected older glibc / libpthread libraries
    // (when launching from inside VS Code snap) which crashes WebKitNetworkProcess.
    char *ld_lib = getenv("LD_LIBRARY_PATH");
    if (ld_lib != NULL && strstr(ld_lib, "/snap/") != NULL) {
        char *dup = strdup(ld_lib);
        if (dup != NULL) {
            char new_ld[8192] = "";
            char *saveptr = NULL;
            char *token = strtok_r(dup, ":", &saveptr);
            while (token != NULL) {
                if (strstr(token, "/snap/") == NULL) {
                    if (new_ld[0] != '\0') {
                        strncat(new_ld, ":", sizeof(new_ld) - strlen(new_ld) - 1);
                    }
                    strncat(new_ld, token, sizeof(new_ld) - strlen(new_ld) - 1);
                }
                token = strtok_r(NULL, ":", &saveptr);
            }
            free(dup);
            if (new_ld[0] == '\0') {
                unsetenv("LD_LIBRARY_PATH");
            } else {
                setenv("LD_LIBRARY_PATH", new_ld, 1);
            }
        }
    }

    // 2. Clear snap-injected preloads & GTK module paths
    char *ld_preload = getenv("LD_PRELOAD");
    if (ld_preload != NULL && strstr(ld_preload, "/snap/") != NULL) {
        unsetenv("LD_PRELOAD");
    }
    char *gtk_path = getenv("GTK_PATH");
    if (gtk_path != NULL && strstr(gtk_path, "/snap/") != NULL) {
        unsetenv("GTK_PATH");
    }
    char *gtk_modules = getenv("GTK_MODULES");
    if (gtk_modules != NULL && strstr(gtk_modules, "/snap/") != NULL) {
        unsetenv("GTK_MODULES");
    }
    char *gio_module_dir = getenv("GIO_MODULE_DIR");
    if (gio_module_dir != NULL && strstr(gio_module_dir, "/snap/") != NULL) {
        unsetenv("GIO_MODULE_DIR");
    }

    // 3. Fix WebKitGTK 2.42+ NVIDIA DMA-BUF black screen bug
    if (getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == NULL) {
        setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1", 1);
    }
}
*/
import "C"

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Ikemen GO Studio",
		Width:     1100,
		Height:    760,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		OnStartup:        app.startup,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon:             icon,
			ProgramName:      "ikemen-studio",
			WebviewGpuPolicy: linux.WebviewGpuPolicyOnDemand,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
