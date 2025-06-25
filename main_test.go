package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "空のファイル",
			data:     []byte{},
			expected: false,
		},
		{
			name:     "テキストファイル",
			data:     []byte("Hello, World!\n"),
			expected: false,
		},
		{
			name:     "NULL文字を含むバイナリファイル",
			data:     []byte{0x00, 0x01, 0x02},
			expected: true,
		},
		{
			name:     "多くの非印字文字を含むファイル",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C},
			expected: true,
		},
		{
			name:     "改行、タブ、キャリッジリターンを含むテキスト",
			data:     []byte("Hello\tWorld\r\nTest"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinary(tt.data)
			if result != tt.expected {
				t.Errorf("isBinary() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "通常のGoファイル",
			path:     "main.go",
			expected: false,
		},
		{
			name:     "隠しファイル",
			path:     ".gitignore",
			expected: true,
		},
		{
			name:     "隠しディレクトリ内のファイル",
			path:     ".git/config",
			expected: true,
		},
		{
			name:     "実行ファイル",
			path:     "program.exe",
			expected: true,
		},
		{
			name:     "画像ファイル",
			path:     "image.jpg",
			expected: true,
		},
		{
			name:     "テキストファイル",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "データベースファイル",
			path:     "data.sqlite",
			expected: true,
		},
		{
			name:     "大文字拡張子",
			path:     "IMAGE.PNG",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipFile(tt.path)
			if result != tt.expected {
				t.Errorf("shouldSkipFile(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCheckAndFixFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "check-new-line-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name            string
		fileContent     string
		fix             bool
		expectedResult  bool
		expectedContent string
		expectedError   bool
	}{
		{
			name:            "改行で終わるファイル",
			fileContent:     "Hello, World!\n",
			fix:             false,
			expectedResult:  true,
			expectedContent: "Hello, World!\n",
			expectedError:   false,
		},
		{
			name:            "改行で終わらないファイル（修正なし）",
			fileContent:     "Hello, World!",
			fix:             false,
			expectedResult:  false,
			expectedContent: "Hello, World!",
			expectedError:   false,
		},
		{
			name:            "改行で終わらないファイル（修正あり）",
			fileContent:     "Hello, World!",
			fix:             true,
			expectedResult:  false,
			expectedContent: "Hello, World!\n",
			expectedError:   false,
		},
		{
			name:            "空のファイル",
			fileContent:     "",
			fix:             false,
			expectedResult:  true,
			expectedContent: "",
			expectedError:   false,
		},
		{
			name:            "バイナリファイル",
			fileContent:     string([]byte{0x00, 0x01, 0x02}),
			fix:             false,
			expectedResult:  true,
			expectedContent: string([]byte{0x00, 0x01, 0x02}),
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストファイルを作成
			testFile := filepath.Join(tempDir, "test_file.txt")
			err := os.WriteFile(testFile, []byte(tt.fileContent), 0o644)
			if err != nil {
				t.Fatalf("テストファイルの作成に失敗: %v", err)
			}

			// 関数をテスト
			result, err := checkAndFixFile(testFile, tt.fix)

			// エラーのチェック
			if tt.expectedError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("予期しないエラーが発生: %v", err)
			}

			// 結果のチェック
			if result != tt.expectedResult {
				t.Errorf("checkAndFixFile() = %v, expected %v", result, tt.expectedResult)
			}

			// ファイル内容のチェック
			actualContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("テストファイルの読み込みに失敗: %v", err)
			}

			if string(actualContent) != tt.expectedContent {
				t.Errorf("ファイル内容が期待値と異なります。actual: %q, expected: %q",
					string(actualContent), tt.expectedContent)
			}
		})
	}
}

func TestProcessRepository(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "check-new-line-repo-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用のファイル構造を作成
	testFiles := map[string]string{
		"file1.txt":       "content without newline",
		"file2.txt":       "content with newline\n",
		"subdir/file3.go": "package main\nfunc main() {}",
		".hidden":         "hidden file",
		"binary.exe":      "binary content",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)

		// ディレクトリを作成
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("ディレクトリの作成に失敗: %v", err)
		}

		// ファイルを作成
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}
	}

	// 修正なしでテスト
	err = processRepository(tempDir, false)
	if err != nil {
		t.Errorf("processRepository()でエラーが発生: %v", err)
	}

	// 修正ありでテスト
	err = processRepository(tempDir, true)
	if err != nil {
		t.Errorf("processRepository()でエラーが発生: %v", err)
	}

	// 修正後、改行が追加されているかを確認
	fixedContent, err := os.ReadFile(filepath.Join(tempDir, "file1.txt"))
	if err != nil {
		t.Fatalf("修正後のファイル読み込みに失敗: %v", err)
	}

	if !strings.HasSuffix(string(fixedContent), "\n") {
		t.Errorf("ファイルが修正されていません。実際の内容: %q", string(fixedContent))
	}
}

func TestProcessRepositoryNonExistentPath(t *testing.T) {
	err := processRepository("/non/existent/path", false)
	if err == nil {
		t.Errorf("存在しないパスに対してエラーが発生しませんでした")
	}
}

// ベンチマークテスト
func BenchmarkIsBinary(b *testing.B) {
	data := []byte("This is a test file with some content that is not binary.\nIt has multiple lines.\nAnd some more content.")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isBinary(data)
	}
}

func BenchmarkShouldSkipFile(b *testing.B) {
	testPaths := []string{
		"main.go",
		".gitignore",
		"image.jpg",
		"document.pdf",
		"script.sh",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			shouldSkipFile(path)
		}
	}
}

// テストヘルパー関数
func createTempFileWithContent(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("一時ファイルの作成に失敗: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("一時ファイルへの書き込みに失敗: %v", err)
	}

	return tempFile.Name()
}

// エラーケーステスト
func TestCheckAndFixFileErrors(t *testing.T) {
	t.Run("存在しないファイル", func(t *testing.T) {
		_, err := checkAndFixFile("/non/existent/file.txt", false)
		if err == nil {
			t.Errorf("存在しないファイルに対してエラーが発生しませんでした")
		}
	})
}

// 統合テスト
func TestIntegration(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 複雑なディレクトリ構造を作成
	structure := map[string]string{
		"src/main.go":      "package main\n\nfunc main() {}",        // 改行なし
		"src/utils.go":     "package main\n\nfunc helper() {}\n",    // 改行あり
		"docs/README.md":   "# Project\n\nDescription",              // 改行なし
		".git/config":      "[core]\n\trepositoryformatversion = 0", // 隠しファイル
		"assets/image.png": "fake png content",                      // バイナリファイル（スキップ対象）
		"data/empty.txt":   "",                                      // 空ファイル
	}

	for path, content := range structure {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("ディレクトリの作成に失敗: %v", err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("ファイルの作成に失敗: %v", err)
		}
	}

	// チェックモードで実行
	err = processRepository(tempDir, false)
	if err != nil {
		t.Errorf("チェックモードでエラー: %v", err)
	}

	// 修正モードで実行
	err = processRepository(tempDir, true)
	if err != nil {
		t.Errorf("修正モードでエラー: %v", err)
	}

	// 修正結果を確認
	checkFiles := []string{"src/main.go", "docs/README.md"}
	for _, file := range checkFiles {
		content, err := os.ReadFile(filepath.Join(tempDir, file))
		if err != nil {
			t.Errorf("修正後のファイル読み込みエラー (%s): %v", file, err)
			continue
		}

		if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
			t.Errorf("ファイル %s が修正されていません", file)
		}
	}
}
