/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2024 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cors

import (
	"bytes"
	"os"
	"testing"
)

func TestCORSXMLMarshal(t *testing.T) {
	fileContents, err := os.ReadFile("testdata/example.xml")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseBucketCorsConfig(bytes.NewReader(fileContents))
	if err != nil {
		t.Fatal(err)
	}
	remarshalled, err := c.ToXML()
	if err != nil {
		t.Fatal(err)
	}

	// 规范化两者：移除所有空白字符差异
	normalize := func(data []byte) string {
		// 移除行首行尾的空白字符
		lines := bytes.Split(data, []byte("\n"))
		var normalized []string
		for _, line := range lines {
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				normalized = append(normalized, string(trimmed))
			}
		}
		// 将所有行连接成一个字符串，用换行符分隔
		return string(bytes.Join([][]byte{
			[]byte(normalized[0]),
			[]byte(normalized[1]),
		}, []byte("\n")))
	}

	got := normalize(remarshalled)
	want := normalize(bytes.TrimSpace(fileContents))

	if got != want {
		t.Errorf("XML mismatch:\ngot:  %s\nwant: %s", got, want)
	}
}
