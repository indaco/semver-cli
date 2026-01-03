package auditlog

import (
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/config"
)

var (
	defaultAuditLog    AuditLog
	RegisterAuditLogFn = registerAuditLog
	GetAuditLogFn      = getAuditLog
)

func registerAuditLog(al AuditLog) {
	if defaultAuditLog != nil {
		fmt.Fprintf(os.Stderr,
			"WARNING: Ignoring audit log %q: another audit log (%q) is already registered.\n",
			al.Name(), defaultAuditLog.Name(),
		)
		return
	}
	defaultAuditLog = al
}

func getAuditLog() AuditLog {
	return defaultAuditLog
}

// ResetAuditLog clears the registered audit log (for testing).
func ResetAuditLog() {
	defaultAuditLog = nil
}

// Register registers the audit log plugin with the given configuration.
func Register(cfg *config.AuditLogConfig) {
	internalCfg := fromConfigStruct(cfg)
	RegisterAuditLogFn(NewAuditLog(internalCfg))
}

// fromConfigStruct converts the config package struct to internal config.
func fromConfigStruct(cfg *config.AuditLogConfig) *Config {
	if cfg == nil {
		return DefaultConfig()
	}

	return &Config{
		Enabled:          cfg.Enabled,
		Path:             cfg.GetPath(),
		Format:           cfg.GetFormat(),
		IncludeAuthor:    cfg.IncludeAuthor,
		IncludeTimestamp: cfg.IncludeTimestamp,
		IncludeCommitSHA: cfg.IncludeCommitSHA,
		IncludeBranch:    cfg.IncludeBranch,
	}
}
