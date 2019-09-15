//  Copyright 2019 Jacques Marais
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package relapseviz

import (
	"fmt"
	"os"
	"testing"
)

// var tt = `(.WhatsUp: == "F" &.Survived: >= 1000000/*years*/ &
// .DragonsExist != true &
// .MonkeysSmart :: $bool &
// .History [*,
// _ == "Katydids Alive"
// ] &
// .FeatureRequests : ._ {Name *= "art";*;
// Anatomy $= "omen";})/*test2*/
// /*test1*/`

var tt = `
(
	.WhatsUp == "E" &
	.Survived >= 1000000 /*years*/ &
	.DragonsExist != true &
	.MonkeysSmart :: $bool &
	.History [
		*,
		_ == "Katydids Alive"
	] &
	.FeatureRequests._ {
		Name *= "art";
		*;
		Anatomy $= "omen";
	} &
	( .WhatsUp: * | .Survived: * | .History._: -> contains($string, "Met" ) )
)
`

func TestTranslate(t *testing.T) {
	graph, err := Translate(tt, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Nodes:\n")
	for n := range graph.Nodes.Lookup {
		fmt.Printf("%v\n", n)
	}
	fmt.Printf("Edges:\n")
	for _, e := range graph.Edges.Edges {
		fmt.Printf("%v -> %v\n", e.Src, e.Dst)
	}
	fmt.Printf("Graph:\n%v\n", graph.String())
	f, _ := os.Create("relapse.svg")
	err = WriteSVG(graph, f)
	if err != nil {
		t.Fatal(err)
	}
}
