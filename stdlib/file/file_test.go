package file

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := "hello world"
	if err := Write(path, content); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != content {
		t.Errorf("Read() = %q, want %q", got, content)
	}
}

func TestReadWriteBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")

	data := []byte{0x00, 0x01, 0x02, 0xFF}
	if err := WriteBytes(path, data); err != nil {
		t.Fatalf("WriteBytes() error = %v", err)
	}

	got, err := ReadBytes(path)
	if err != nil {
		t.Fatalf("ReadBytes() error = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadBytes() = %v, want %v", got, data)
	}
}

func TestReadWriteLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")

	lines := []string{"line 1", "line 2", "line 3"}
	if err := WriteLines(path, lines); err != nil {
		t.Fatalf("WriteLines() error = %v", err)
	}

	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(got) != len(lines) {
		t.Errorf("ReadLines() len = %d, want %d", len(got), len(lines))
	}
	for i := range lines {
		if got[i] != lines[i] {
			t.Errorf("ReadLines()[%d] = %q, want %q", i, got[i], lines[i])
		}
	}
}

func TestAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.txt")

	if err := Write(path, "hello"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Append(path, " world"); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != "hello world" {
		t.Errorf("Read() = %q, want %q", got, "hello world")
	}
}

func TestAppendCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.txt")

	if err := Append(path, "created by append"); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != "created by append" {
		t.Errorf("Read() = %q, want %q", got, "created by append")
	}
}

func TestAppendBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.bin")

	if err := WriteBytes(path, []byte{0x01}); err != nil {
		t.Fatalf("WriteBytes() error = %v", err)
	}
	if err := AppendBytes(path, []byte{0x02}); err != nil {
		t.Fatalf("AppendBytes() error = %v", err)
	}

	got, err := ReadBytes(path)
	if err != nil {
		t.Fatalf("ReadBytes() error = %v", err)
	}
	if len(got) != 2 || got[0] != 0x01 || got[1] != 0x02 {
		t.Errorf("ReadBytes() = %v, want [1 2]", got)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")

	if Exists(path) {
		t.Error("Exists() = true for non-existent file")
	}

	if err := Write(path, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if !Exists(path) {
		t.Error("Exists() = false for existing file")
	}
}

func TestSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "size.txt")

	content := "12345"
	if err := Write(path, content); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	size, err := Size(path)
	if err != nil {
		t.Fatalf("Size() error = %v", err)
	}
	if size != int64(len(content)) {
		t.Errorf("Size() = %d, want %d", size, len(content))
	}
}

func TestDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := Write(path, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if Directory(path) {
		t.Error("Directory() = true for file")
	}
	if !Directory(dir) {
		t.Error("Directory() = false for directory")
	}
}

func TestFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := Write(path, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if !File(path) {
		t.Error("File() = false for file")
	}
	if File(dir) {
		t.Error("File() = true for directory")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remove.txt")

	if err := Write(path, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if Exists(path) {
		t.Error("file still exists after Remove()")
	}
}

func TestRemoveAll(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	path := filepath.Join(subdir, "file.txt")

	if err := MkdirAll(subdir); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := Write(path, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := RemoveAll(subdir); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if Exists(subdir) {
		t.Error("directory still exists after RemoveAll()")
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")

	if err := Write(oldPath, "test"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Rename(oldPath, newPath); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if Exists(oldPath) {
		t.Error("old file still exists after Rename()")
	}
	if !Exists(newPath) {
		t.Error("new file doesn't exist after Rename()")
	}
}

func TestCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	content := "copy me"
	if err := Write(src, content); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := Copy(src, dst); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	got, err := Read(dst)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != content {
		t.Errorf("Read() = %q, want %q", got, content)
	}
}

func TestMkdir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newdir")

	if err := Mkdir(path); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if !Directory(path) {
		t.Error("Directory() = false after Mkdir()")
	}
}

func TestMkdirAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c")

	if err := MkdirAll(path); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if !Directory(path) {
		t.Error("Directory() = false after MkdirAll()")
	}
}

func TestReadNonExistent(t *testing.T) {
	_, err := Read("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Read() should return error for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Read() error = %v, want os.ErrNotExist", err)
	}
}

func TestSizeNonExistent(t *testing.T) {
	_, err := Size("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Size() should return error for non-existent file")
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	if err := WriteLines(path, []string{}); err != nil {
		t.Fatalf("WriteLines() error = %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != "" {
		t.Errorf("Read() = %q, want empty string", got)
	}
}
