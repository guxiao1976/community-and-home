package wordstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
	"github.com/guxiao/community-and-home/services/moderation/internal/normalize"
	"github.com/guxiao/community-and-home/services/moderation/internal/pinyin"
	"github.com/guxiao/community-and-home/services/moderation/internal/splitword"
	"github.com/guxiao/community-and-home/services/moderation/internal/whitelist"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type WordStore struct {
	db            sqlx.SqlConn
	acMachine     *ac.ACMachine
	wl            *whitelist.Whitelist
	version       int64
	syncInterval  int
	pinyinExpand  bool
	pinyinMax     int
	splitSeparators string
}

type WordStoreOption func(*WordStore)

func WithSyncInterval(seconds int) WordStoreOption {
	return func(ws *WordStore) { ws.syncInterval = seconds }
}

func WithPinyinExpand(enable bool, maxVariants int) WordStoreOption {
	return func(ws *WordStore) { ws.pinyinExpand = enable; ws.pinyinMax = maxVariants }
}

func WithSplitSeparators(sep string) WordStoreOption {
	return func(ws *WordStore) { ws.splitSeparators = sep }
}

func NewWordStore(db sqlx.SqlConn, opts ...WordStoreOption) *WordStore {
	ws := &WordStore{
		db:             db,
		acMachine:      ac.NewACMachine(),
		wl:             whitelist.NewWhitelist(),
		syncInterval:   300,
		pinyinExpand:   true,
		pinyinMax:      20,
		splitSeparators: "xX.* 、\t-_|/~·丶",
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

func (s *WordStore) GetACMachine() *ac.ACMachine     { return s.acMachine }
func (s *WordStore) GetWhitelist() *whitelist.Whitelist { return s.wl }

func (s *WordStore) Load(ctx context.Context) error {
	return s.load(ctx, true)
}

func (s *WordStore) Reload(ctx context.Context) error {
	return s.load(ctx, false)
}

func (s *WordStore) load(ctx context.Context, initial bool) error {
	// Check version
	if !initial {
		newVer, err := s.getVersion(ctx)
		if err != nil {
			return fmt.Errorf("check version failed: %w", err)
		}
		if newVer <= s.version {
			return nil
		}
		logx.Infof("wordstore: version changed %d → %d, rebuilding", s.version, newVer)
	}

	// Load blacklist (word_type=1)
	blackEntries, err := s.loadWords(ctx, 1)
	if err != nil {
		return fmt.Errorf("load blacklist failed: %w", err)
	}

	// Expand with pinyin homophones for high-risk words
	if s.pinyinExpand {
		expanded := ac.ExpandEntriesWithPinyin(blackEntries, pinyin.ExpandHomophones, s.pinyinMax)
		blackEntries = expanded
	}

	if err := s.acMachine.Build(blackEntries); err != nil {
		return fmt.Errorf("build ac machine failed: %w", err)
	}

	// Load whitelist (word_type=2)
	allowWords, err := s.loadRawWords(ctx, 2)
	if err != nil {
		return fmt.Errorf("load whitelist failed: %w", err)
	}
	s.wl.Build(allowWords)

	// Update version
	ver, err := s.getVersion(ctx)
	if err != nil {
		return fmt.Errorf("get version failed: %w", err)
	}
	s.version = ver

	logx.Infof("wordstore: loaded %d blacklist words, %d whitelist words, version=%d",
		len(blackEntries), len(allowWords), s.version)
	return nil
}

func (s *WordStore) loadWords(ctx context.Context, wordType int64) ([]ac.WordEntry, error) {
	query := `SELECT word, category, severity FROM md_sensitive_word WHERE status = 1 AND word_type = ? AND delete_time IS NULL`
	var entries []ac.WordEntry
	err := s.db.QueryRowsCtx(ctx, &entries, query, wordType)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if entries == nil {
		entries = []ac.WordEntry{}
	}
	return entries, nil
}

func (s *WordStore) loadRawWords(ctx context.Context, wordType int64) ([]string, error) {
	query := `SELECT word FROM md_sensitive_word WHERE status = 1 AND word_type = ? AND delete_time IS NULL`
	var rows []struct {
		Word string `db:"word"`
	}
	err := s.db.QueryRowsCtx(ctx, &rows, query, wordType)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	words := make([]string, 0, len(rows))
	for _, r := range rows {
		words = append(words, r.Word)
	}
	return words, nil
}

func (s *WordStore) getVersion(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(MAX(UNIX_TIMESTAMP(updated_time)), 0) FROM md_sensitive_word WHERE status = 1 AND delete_time IS NULL`
	var ver int64
	err := s.db.QueryRowCtx(ctx, &ver, query)
	if err != nil {
		return 0, err
	}
	return ver, nil
}

// GetSplitDetector creates a new SplitDetector from config
func (s *WordStore) GetSplitDetector() *splitword.SplitDetector {
	return splitword.NewSplitDetector(s.splitSeparators)
}

// GetNormalizer creates a new Normalizer from default config
func (s *WordStore) GetNormalizer() *normalize.Normalizer {
	return normalize.New(
		normalize.WithWidth(),
		normalize.WithCase(),
		normalize.WithChinese(),
		normalize.WithNumStyle(),
		normalize.WithEnglishStyle(),
		normalize.WithRepeat(),
	)
}

func (s *WordStore) StartSync(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Duration(s.syncInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logx.Info("wordstore: sync stopped")
				return
			case <-ticker.C:
				if err := s.Reload(ctx); err != nil {
					logx.Errorf("wordstore: reload failed: %v", err)
				}
			}
		}
	}()
	logx.Infof("wordstore: sync started, interval=%ds", s.syncInterval)
}
