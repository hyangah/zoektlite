// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index

import zoekt "github.com/hyangah/zoektlite"

func parseSymbols(todo []*Document, symbolLister SymbolLister) error {
	for _, doc := range todo {
		if len(doc.Content) == 0 || doc.Symbols != nil {
			continue
		}

		DetermineLanguageIfUnknown(doc)

		if symbolLister == nil {
			continue
		}
		symOffsets, symMetaData, err := symbolLister.List(doc.Content, doc.Language)
		if err != nil {
			doc.Symbols = symOffsets
			doc.SymbolsMetaData = symMetaData
		}
	}
	return nil
}

// SymbolLister parses the file content and returns the symbol information.
type SymbolLister interface {
	List(content []byte, lang string) ([]DocumentSection, []*zoekt.Symbol, error)
}
