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
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/jmarais/relapseviz/svg"
	"github.com/katydid/katydid/relapse"
	"github.com/katydid/katydid/relapse/ast"
	"github.com/katydid/katydid/relapse/types"
)

type translator struct {
	graph *gographviz.Graph
	full  bool
	r     *rand.Rand
}

func Translate(s string, full bool) (*gographviz.Graph, error) {
	g, err := relapse.Parse(s)
	if err != nil {
		return nil, err
	}
	return TranslateGrammar(g, full), nil
}
func TranslateGrammar(g *ast.Grammar, full bool) *gographviz.Graph {
	t := &translator{
		graph: gographviz.NewGraph(),
		full:  full,
		r:     rand.New(rand.NewSource(0)),
	}
	if err := t.graph.SetName("Relapse"); err != nil {
		panic(err)
	}
	if err := t.graph.SetDir(true); err != nil {
		panic(err)
	}
	nodeLabel := getTypeName(g)
	t.translate(g, "root"+nodeLabel)
	return t.graph
}

// Get the ast type name
func getTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	typeName := rv.Type().String()
	ss := strings.Split(typeName, ".")
	return ss[len(ss)-1]
}

func (t *translator) translate(node interface{}, nodeId string) {
	switch v := node.(type) {
	case *ast.Grammar:
		label := newLabel(getTypeName(v))
		if v.After != nil {
			label.write(`\nAfter: \"`, v.After.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.After != nil {
				t.down(nodeId, v.After)
			}
		}
		if v.TopPattern != nil {
			t.down(nodeId, v.TopPattern)
		}
		for i, pdecl := range v.PatternDecls {
			nextNodeName := getTypeName(pdecl)
			nextNodeId := nextNodeName + strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeId, map[string]string{attrLabel: fmt.Sprintf("\"%v[%d]\"", nextNodeName, i)})
			t.translate(pdecl, nextNodeId)
		}
	case *ast.PatternDecl:
		label := newLabel(getTypeName(v))
		if v.Name != "" {
			label.write(`\nName: `, v.Name)
		}
		if v.Hash != nil {
			label.write(`\nHash: `, v.Hash.String())
		}
		if v.Eq != nil {
			label.write(`\nEq: `, v.Eq.String())
		}
		if v.Before != nil {
			label.write(`\nBefore: `, v.Before.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Hash != nil {
				t.down(nodeId, v.Hash)
			}
			if v.Eq != nil {
				t.down(nodeId, v.Eq)
			}
			if v.Before != nil {
				t.down(nodeId, v.Before)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.Pattern:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Empty != nil {
			t.down(nodeId, v.Empty)
		}
		if v.TreeNode != nil {
			t.down(nodeId, v.TreeNode)
		}
		if v.LeafNode != nil {
			t.down(nodeId, v.LeafNode)
		}
		if v.Concat != nil {
			t.down(nodeId, v.Concat)
		}
		if v.Or != nil {
			t.down(nodeId, v.Or)
		}
		if v.And != nil {
			t.down(nodeId, v.And)
		}
		if v.ZeroOrMore != nil {
			t.down(nodeId, v.ZeroOrMore)
		}
		if v.Reference != nil {
			t.down(nodeId, v.Reference)
		}
		if v.Not != nil {
			t.down(nodeId, v.Not)
		}
		if v.ZAny != nil {
			t.down(nodeId, v.ZAny)
		}
		if v.Contains != nil {
			t.down(nodeId, v.Contains)
		}
		if v.Optional != nil {
			t.down(nodeId, v.Optional)
		}
		if v.Interleave != nil {
			t.down(nodeId, v.Interleave)
		}
	case *ast.Empty:
		label := newLabel(getTypeName(v))
		if v.Empty != nil {
			label.write(`\nEmpty: \"`, v.Empty.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Empty != nil {
				t.down(nodeId, v.Empty)
			}
		}
	case *ast.TreeNode:
		label := newLabel(getTypeName(v))
		if v.Name != nil {
			label.write(`\nName: `, v.Name.String())
		}
		if v.Colon != nil {
			label.write(`\nColon: `, v.Colon.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Colon != nil {
				t.down(nodeId, v.Colon)
			}
		}
		if v.Name != nil {
			t.down(nodeId, v.Name)
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.LeafNode:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Expr != nil {
			t.down(nodeId, v.Expr)
		}
	case *ast.Concat:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern)
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern)
		}
	case *ast.Or:
		label := newLabel(getTypeName(v))
		if v.Pipe != nil {
			label.write(`\nPipe: `, v.Pipe.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern)
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern)
		}
	case *ast.And:
		label := newLabel(getTypeName(v))
		if v.Ampersand != nil {
			label.write(`\nAmpersand: `, v.Ampersand.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern)
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern)
		}
	case *ast.ZeroOrMore:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.Reference:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.Not:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.ZAny:
		label := newLabel(getTypeName(v))
		if v.Star != nil {
			label.write(`\nStar: `, v.Star.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Star != nil {
				t.down(nodeId, v.Star)
			}
		}
	case *ast.Contains:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.Optional:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern)
		}
	case *ast.Interleave:
		label := newLabel(getTypeName(v))
		label.write(`\n`)
		if v.OpenCurly != nil {
			label.write(v.OpenCurly.String())
		}
		if v.LeftPattern != nil {
			label.write(`Left`)
		}
		if v.SemiColon != nil {
			label.write(v.SemiColon.String())
		}
		if v.SemiColon != nil {
			label.write(`Right`)
		}
		if v.ExtraSemiColon != nil {
			label.write(v.ExtraSemiColon.String())
		}
		if v.CloseCurly != nil {
			label.write(v.CloseCurly.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern)
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern)
		}
	case *ast.Expr:
		label := newLabel(getTypeName(v))
		if v.RightArrow != nil {
			label.write(`\nRightArrow: `, v.RightArrow.String())
		}
		if v.Comma != nil {
			label.write(`\nComma: `, v.Comma.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Terminal != nil {
			t.down(nodeId, v.Terminal)
		}
		if v.List != nil {
			t.down(nodeId, v.List)
		}
		if v.Function != nil {
			t.down(nodeId, v.Function)
		}
		if v.BuiltIn != nil {
			t.down(nodeId, v.BuiltIn)
		}
	case *ast.NameExpr:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Name != nil {
			t.down(nodeId, v.Name)
		}
		if v.AnyName != nil {
			t.down(nodeId, v.AnyName)
		}
		if v.AnyNameExcept != nil {
			t.down(nodeId, v.AnyNameExcept)
		}
		if v.NameChoice != nil {
			t.down(nodeId, v.NameChoice)
		}
	case *ast.Name:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.AnyName:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.AnyNameExcept:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Except != nil {
			t.down(nodeId, v.Except)
		}
	case *ast.NameChoice:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Left != nil {
			t.down(nodeId, v.Left)
		}
		if v.Right != nil {
			t.down(nodeId, v.Right)
		}
	case *ast.List:
		label := newLabel(getTypeName(v))
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		for i, e := range v.GetElems() {
			nextNodeName := getTypeName(e)
			nextNodeId := nextNodeName + strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeId, map[string]string{attrLabel: fmt.Sprintf("\"%v[%d]\"", nextNodeName, i)})
			t.translate(e, nextNodeId)
		}
	case *ast.Function:
		label := newLabel(getTypeName(v))
		if v.Name != "" {
			label.write(`\nName: `, v.Name)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		for i, e := range v.GetParams() {
			nextNodeName := getTypeName(e)
			nextNodeId := nextNodeName + strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeId, map[string]string{attrLabel: fmt.Sprintf("\"%v[%d]\"", nextNodeName, i)})
			t.translate(e, nextNodeId)
		}
	case *ast.BuiltIn:
		label := newLabel(getTypeName(v))
		if v.Symbol != nil {
			label.write(`\nSymbol: `, v.Symbol.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Expr != nil {
			t.down(nodeId, v.Expr)
		}
	case *ast.Terminal:
		label := newLabel(getTypeName(v))
		if v.Literal != "" {
			label.write(`\nLiteral: `, strings.Replace(v.Literal, "\"", "\\\"", -1))
		}
		if v.DoubleValue != nil {
			label.write(`\nDoubleValue: `, strconv.FormatFloat(*v.DoubleValue, 'E', -1, 64))
		}
		if v.IntValue != nil {
			label.write(`\nIntValue: `, strconv.FormatInt(*v.IntValue, 10))
		}
		if v.UintValue != nil {
			label.write(`\nUintValue: `, strconv.FormatUint(*v.UintValue, 10))
		}
		if v.BoolValue != nil {
			label.write(`\nBoolValue: `, strconv.FormatBool(*v.BoolValue))
		}
		if v.StringValue != nil {
			label.write(`\nStringValue: `, *v.StringValue)
		}
		if v.BytesValue != nil {
			label.write(`\nBytesValue: `, string(v.BytesValue))
		}
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Variable != nil {
			t.down(nodeId, v.Variable)
		}
	case *ast.Variable:
		name := types.Type_name[int32(v.Type)]
		label := newLabel(getTypeName(v))
		label.write(`:\nType:`, name)
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.Space:
		label := newLabel(getTypeName(v))
		for i, s := range v.Space {
			label.write(`\nSpace[`, strconv.Itoa(i), `]: \"`, s, `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.Keyword:
		label := newLabel(getTypeName(v))
		if v.Value != "" {
			label.write(`\nValue: \"`, v.Value, `\"`)
		}
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Before != nil {
			t.down(nodeId, v.Before)
		}
	}

	fmt.Printf("Unknown\n")
}

var attrLabel = string(gographviz.Label)

func (t *translator) down(nodeId string, to interface{}) {
	nextNodeName := getTypeName(to)
	nextNodeId := nextNodeName + strconv.FormatUint(t.r.Uint64(), 10)
	t.addEdge(nodeId, nextNodeId, map[string]string{attrLabel: nextNodeName})
	t.translate(to, nextNodeId)
}

func (t *translator) addNode(name string, attr map[string]string) {
	if err := t.graph.AddNode(t.graph.Name, name, attr); err != nil {
		panic(err)
	}
}

func (t *translator) addEdge(from, to string, attr map[string]string) {
	if err := t.graph.AddEdge(from, to, true, attr); err != nil {
		panic(err)
	}
}

type label struct {
	b *strings.Builder
}

func newLabel(name string) *label {
	b := &strings.Builder{}
	b.WriteString(`"`)
	b.WriteString(name)
	return &label{b}
}

func (l *label) write(ss ...string) {
	for _, s := range ss {
		l.b.WriteString(s)
	}
}

func (l *label) finish() string {
	l.b.WriteString(`"`)
	return l.b.String()
}

func WriteSVG(graph *gographviz.Graph, w io.Writer) error {
	pp := svg.MassageDotSVG()
	return pp(bytes.NewReader([]byte(graph.String())), w)
}
