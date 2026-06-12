package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)      // Linux/macOS
	t.Setenv("USERPROFILE", tmp) // Windows
}

func sampleRecord(t time.Time) Record {
	return Record{
		Timestamp:    t,
		Server:       "ndt-mlab1-sin01.mlab-oti.measurement-lab.org",
		Location:     "Singapore SG",
		ISP:          "Maxis",
		LatencyMs:    14.2,
		JitterMs:     0.9,
		DownloadMbps: 94.71,
		UploadMbps:   23.10,
		DataUsedMB:   44.9,
	}
}

func TestSaveAndLoad(t *testing.T) {
	withTempHome(t)

	r1 := sampleRecord(time.Now().Add(-2 * time.Hour))
	r2 := sampleRecord(time.Now().Add(-1 * time.Hour))
	r3 := sampleRecord(time.Now())

	require.NoError(t, Save(r1))
	require.NoError(t, Save(r2))
	require.NoError(t, Save(r3))

	records, err := Load(0)
	require.NoError(t, err)
	assert.Len(t, records, 3)
	assert.Equal(t, "Maxis", records[0].ISP)
}

func TestLoad_LimitN(t *testing.T) {
	withTempHome(t)

	for i := 0; i < 5; i++ {
		require.NoError(t, Save(sampleRecord(time.Now())))
	}

	records, err := Load(3)
	require.NoError(t, err)
	assert.Len(t, records, 3)
}

func TestLoad_EmptyFile(t *testing.T) {
	withTempHome(t)

	records, err := Load(10)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestLoad_MalformedLines(t *testing.T) {
	withTempHome(t)

	path, err := Path()
	require.NoError(t, err)

	// Write one good and one bad line.
	good, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	good.WriteString("{\"timestamp\":\"2026-01-01T00:00:00Z\",\"server\":\"test\"}\n")
	good.WriteString("not json\n")
	good.Close()

	records, err := Load(0)
	require.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestPath_CreatesDirectory(t *testing.T) {
	withTempHome(t)

	p, err := Path()
	require.NoError(t, err)
	assert.Equal(t, ".speeder/history.jsonl", filepath.Base(filepath.Dir(p))+"/"+filepath.Base(p))
	assert.DirExists(t, filepath.Dir(p))
}

func TestSave_PingOnly(t *testing.T) {
	withTempHome(t)

	r := sampleRecord(time.Now())
	r.PingOnly = true
	r.DownloadMbps = 0
	r.UploadMbps = 0

	require.NoError(t, Save(r))

	records, err := Load(1)
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.True(t, records[0].PingOnly)
}
