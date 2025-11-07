package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
    "github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
    "github.com/tuusuario/afs-challenge/internal/usecases/validation"
)

func main() {
    // flags
    var wait bool
    flag.BoolVar(&wait, "wait", true, "wait for validation to complete (default true)")
    flag.Parse()

    // logger
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    fmt.Println("\n📋 === PITR Validation Started ===")
    fmt.Println("⏰ Timestamp:", time.Now().Format(time.RFC3339))
    fmt.Println("🔧 Loading configuration...")

    cfg, err := cfgpkg.Load()
    if err != nil {
        log.Fatalf("❌ config load: %v", err)
    }
    fmt.Println("✅ Configuration loaded")
    fmt.Printf("   Service: %s\n", cfg.TigerCloud.MainService)
    fmt.Printf("   Project: %s\n", cfg.TigerCloud.ProjectID)

    if !cfg.TigerCloud.UseTigerCloud {
        log.Fatalf("❌ USE_TIGER_CLOUD must be true in environment")
    }

    // create MCP client
    fmt.Println("\n🔌 Initializing MCP Client...")
    httpClient := &http.Client{Timeout: 30 * time.Second}
    client, err := mcp.New(cfg, httpClient)
    if err != nil {
        log.Fatalf("❌ mcp client init: %v", err)
    }
    fmt.Println("✅ MCP Client created")

    fmt.Println("\n🔐 Authenticating with Tiger Cloud...")
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("❌ mcp connect: %v", err)
    }
    fmt.Println("✅ Authentication successful")

    if !wait {
        fmt.Println("\n⏭️  Non-wait mode enabled; exiting")
        os.Exit(0)
    }

    fmt.Println("\n🚀 Starting validation workflow...")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

    res, err := validation.ValidateForksAndPITR(ctx, cfg, client)

    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

    // print structured result even if error
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")

    fmt.Println("\n📊 === VALIDATION RESULT ===")
    if res != nil {
        _ = enc.Encode(res)
    }

    if err != nil {
        fmt.Printf("\n❌ Validation failed: %v\n", err)
        log.Fatalf("Error: %v", err)
    }

    fmt.Println("\n✅ === VALIDATION COMPLETED SUCCESSFULLY ===\n")
}
