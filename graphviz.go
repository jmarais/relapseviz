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

// Translate the given ast.Grammer to a graphviz Graph.
// 'full' traverse extra nodes which the relapse walker skips.
// The node names are generated from the ast type name while a edge
// name will be the fieldname of the edge source.
// The list of struct fields are also listed in the node under the name.
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
	t.translate(g, nodeLabel, `root`)
	return t.graph
}

// Get the ast type name
func getTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	typeName := rv.Type().String()
	ss := strings.Split(typeName, ".")
	return ss[len(ss)-1]
}

func (t *translator) translate(node interface{}, nodeName, suffix string) {
	nodeId := nodeName + suffix
	label := newLabel(nodeName)
	switch v := node.(type) {
	case *ast.Grammar:
		if v.After != nil {
			label.write(`\nAfter: \"`, v.After.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.TopPattern != nil {
			t.down(nodeId, v.TopPattern, `TopPattern`)
		}
		for i, pdecl := range v.PatternDecls {
			nextNodeName := getTypeName(pdecl)
			suff := strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeName+suff, map[string]string{attrLabel: fmt.Sprintf(`"%v[%d]"`, `PatternDecls`, i)})
			t.translate(pdecl, nextNodeName, suff)
		}
		if t.full {
			if v.After != nil {
				t.down(nodeId, v.After, `After`)
			}
		}
	case *ast.PatternDecl:
		if v.Hash != nil {
			label.write(`\nHash: `, v.Hash.String())
		}
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		if v.Name != "" {
			label.write(`\nName: `, v.Name)
		}
		if v.Eq != nil {
			label.write(`\nEq: `, v.Eq.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})

		if t.full {
			if v.Hash != nil {
				t.down(nodeId, v.Hash, `Hash`)
			}
			if v.Eq != nil {
				t.down(nodeId, v.Eq, `Eq`)
			}
			if v.Before != nil {
				t.down(nodeId, v.Before, `Before`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
	case *ast.Pattern:
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Empty != nil {
			t.down(nodeId, v.Empty, `Empty`)
		}
		if v.TreeNode != nil {
			t.down(nodeId, v.TreeNode, `TreeNode`)
		}
		if v.LeafNode != nil {
			t.down(nodeId, v.LeafNode, `LeafNode`)
		}
		if v.Concat != nil {
			t.down(nodeId, v.Concat, `Concat`)
		}
		if v.Or != nil {
			t.down(nodeId, v.Or, `Or`)
		}
		if v.And != nil {
			t.down(nodeId, v.And, `And`)
		}
		if v.ZeroOrMore != nil {
			t.down(nodeId, v.ZeroOrMore, `ZeroOrMore`)
		}
		if v.Reference != nil {
			t.down(nodeId, v.Reference, `Reference`)
		}
		if v.Not != nil {
			t.down(nodeId, v.Not, `Not`)
		}
		if v.ZAny != nil {
			t.down(nodeId, v.ZAny, `ZAny`)
		}
		if v.Contains != nil {
			t.down(nodeId, v.Contains, `Contains`)
		}
		if v.Optional != nil {
			t.down(nodeId, v.Optional, `Optional`)
		}
		if v.Interleave != nil {
			t.down(nodeId, v.Interleave, `Interleave`)
		}
	case *ast.Empty:
		if v.Empty != nil {
			label.write(`\nEmpty: \"`, v.Empty.String(), `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Empty != nil {
				t.down(nodeId, v.Empty, `Empty`)
			}
		}
	case *ast.TreeNode:
		if v.Name != nil {
			label.write(`\nName: `, v.Name.String())
		}
		if v.Colon != nil {
			label.write(`\nColon: `, v.Colon.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Name != nil {
			t.down(nodeId, v.Name, `Name`)
		}
		if t.full {
			if v.Colon != nil {
				t.down(nodeId, v.Colon, `Colon`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
	case *ast.Contains:
		if v.Dot != nil {
			label.write(`\nDot: `, v.Dot.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Dot != nil {
				t.down(nodeId, v.Dot, `Dot`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
	case *ast.LeafNode:
		if v.Expr != nil {
			label.write(`\nExpr: `, `Expr`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Expr != nil {
			t.down(nodeId, v.Expr, `Expr`)
		}
	case *ast.Concat:
		if v.OpenBracket != nil {
			label.write(`\nOpenBracket: `, v.OpenBracket.String())
		}
		if v.LeftPattern != nil {
			label.write(`\nLeftPattern: `, `LeftPattern`)
		}
		if v.Comma != nil {
			label.write(`\nComma: `, v.Comma.String())
		}
		if v.RightPattern != nil {
			label.write(`\nRightPattern: `, `RightPattern`)
		}
		if v.ExtraComma != nil {
			label.write(`\nComma: `, v.ExtraComma.String())
		}
		if v.OpenBracket != nil {
			label.write(`\nCloseBracket: `, v.CloseBracket.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenBracket != nil {
				t.down(nodeId, v.OpenBracket, `OpenBracket`)
			}
		}
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern, `LeftPattern`)
		}
		if t.full {
			if v.Comma != nil {
				t.down(nodeId, v.Comma, `Comma`)
			}
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern, `RightPattern`)
		}
		if t.full {
			if v.ExtraComma != nil {
				t.down(nodeId, v.ExtraComma, `ExtraComma`)
			}
			if v.CloseBracket != nil {
				t.down(nodeId, v.CloseBracket, `CloseBracket`)
			}
		}
	case *ast.Or:
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.LeftPattern != nil {
			label.write(`\nLeftPattern: `, `LeftPattern`)
		}
		if v.Pipe != nil {
			label.write(`\nPipe: `, v.Pipe.String())
		}
		if v.RightPattern != nil {
			label.write(`\nRightPattern: `, `RightPattern`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern, `LeftPattern`)
		}
		if t.full {
			if v.Pipe != nil {
				t.down(nodeId, v.Pipe, `Pipe`)
			}
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern, `RightPattern`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	case *ast.And:
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.LeftPattern != nil {
			label.write(`\nLeftPattern: `, `LeftPattern`)
		}
		if v.Ampersand != nil {
			label.write(`\nAmpersand: `, v.Ampersand.String())
		}
		if v.RightPattern != nil {
			label.write(`\nRightPattern: `, `RightPattern`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern, `LeftPattern`)
		}
		if t.full {
			if v.Ampersand != nil {
				t.down(nodeId, v.Ampersand, `Ampersand`)
			}
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern, `RightPattern`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	case *ast.ZeroOrMore:
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		if v.Star != nil {
			label.write(`\nStar: `, v.Star.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
			if v.Star != nil {
				t.down(nodeId, v.Star, `Star`)
			}
		}
	case *ast.Reference:
		if v.At != nil {
			label.write(`\nAt: `, v.At.String())
		}
		if v.Name != "" {
			label.write(`\nName: `, v.Name)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.At != nil {
				t.down(nodeId, v.At, `At`)
			}
		}
	case *ast.Not:
		if v.Exclamation != nil {
			label.write(`\nExclamation: `, v.Exclamation.String())
		}
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Exclamation != nil {
				t.down(nodeId, v.Exclamation, `Exclamation`)
			}
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	case *ast.ZAny:
		if v.Star != nil {
			label.write(`\nStar: `, v.Star.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Star != nil {
				t.down(nodeId, v.Star, `Star`)
			}
		}
	case *ast.Optional:
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Pattern != nil {
			label.write(`\nPattern: `, `Pattern`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		if v.QuestionMark != nil {
			label.write(`\nQuestionMark: `, v.QuestionMark.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
		if t.full {

			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.Pattern != nil {
			t.down(nodeId, v.Pattern, `Pattern`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
			if v.QuestionMark != nil {
				t.down(nodeId, v.QuestionMark, `QuestionMark`)
			}
		}
	case *ast.Interleave:
		if v.OpenCurly != nil {
			label.write(`\nOpenCurly: `, v.OpenCurly.String())
		}
		if v.LeftPattern != nil {
			label.write(`\nLeftPattern: `, `LeftPattern`)
		}
		if v.SemiColon != nil {
			label.write(`\nSemiColon: `, v.SemiColon.String())
		}
		if v.RightPattern != nil {
			label.write(`\nRightPattern: `, `RightPattern`)
		}
		if v.ExtraSemiColon != nil {
			label.write(`\nExtraSemiColon: `, v.ExtraSemiColon.String())
		}
		if v.CloseCurly != nil {
			label.write(`\nCloseCurly: `, v.CloseCurly.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenCurly != nil {
				t.down(nodeId, v.OpenCurly, `OpenCurly`)
			}
		}
		if v.LeftPattern != nil {
			t.down(nodeId, v.LeftPattern, `LeftPattern`)
		}
		if t.full {
			if v.SemiColon != nil {
				t.down(nodeId, v.SemiColon, `SemiColon`)
			}
		}
		if v.RightPattern != nil {
			t.down(nodeId, v.RightPattern, `RightPattern`)
		}
		if t.full {
			if v.ExtraSemiColon != nil {
				t.down(nodeId, v.ExtraSemiColon, `ExtraSemiColon`)
			}
			if v.CloseCurly != nil {
				t.down(nodeId, v.CloseCurly, `CloseCurly`)
			}
		}
	case *ast.Expr:
		if v.RightArrow != nil {
			label.write(`\nRightArrow: `, v.RightArrow.String())
		}
		if v.Comma != nil {
			label.write(`\nComma: `, v.Comma.String())
		}
		if v.Terminal != nil {
			label.write(`\nTerminal: `, `Terminal`)
		}
		if v.List != nil {
			label.write(`\nList: `, `List`)
		}
		if v.Function != nil {
			label.write(`\nFunction: `, `Function`)
		}
		if v.BuiltIn != nil {
			label.write(`\nBuiltIn: `, `BuiltIn`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.RightArrow != nil {
				t.down(nodeId, v.RightArrow, `RightArrow`)
			}
			if v.Comma != nil {
				t.down(nodeId, v.Comma, `Comma`)
			}
		}
		if v.Terminal != nil {
			t.down(nodeId, v.Terminal, `Terminal`)
		}
		if v.List != nil {
			t.down(nodeId, v.List, `List`)
		}
		if v.Function != nil {
			t.down(nodeId, v.Function, `Function`)
		}
		if v.BuiltIn != nil {
			t.down(nodeId, v.BuiltIn, `BuiltIn`)
		}
	case *ast.List:
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		label.write(`\nType: `, types.Type_name[int32(v.Type)])
		if v.OpenCurly != nil {
			label.write(`\nOpenCurly: `, v.OpenCurly.String())
		}
		if v.Elems != nil {
			label.write(`\nElems: `, `Elems`)
		}
		if v.CloseCurly != nil {
			label.write(`\nCloseCurly: `, v.CloseCurly.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Before != nil {
				t.down(nodeId, v.Before, `Before`)
			}
			if v.OpenCurly != nil {
				t.down(nodeId, v.OpenCurly, `OpenCurly`)
			}
		}
		for i, e := range v.GetElems() {
			nextNodeName := getTypeName(e)
			suff := strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeName+suff, map[string]string{attrLabel: fmt.Sprintf(`"%v[%d]"`, `Elems`, i)})
			t.translate(e, nextNodeName, suff)
		}
		if t.full {
			if v.CloseCurly != nil {
				t.down(nodeId, v.CloseCurly, `CloseCurly`)
			}
		}
	case *ast.Function:
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		if v.Name != "" {
			label.write(`\nName: `, v.Name)
		}
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Params != nil {
			label.write(`\nParams: `, `Params`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Before != nil {
				t.down(nodeId, v.Before, `Before`)
			}
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		for i, e := range v.GetParams() {
			nextNodeName := getTypeName(e)
			suff := strconv.FormatUint(t.r.Uint64(), 10)
			t.addEdge(nodeId, nextNodeName+suff, map[string]string{attrLabel: fmt.Sprintf(`"%v[%d]"`, `Params`, i)})
			t.translate(e, nextNodeName, suff)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	case *ast.BuiltIn:
		if v.Symbol != nil {
			label.write(`\nSymbol: `, v.Symbol.String())
		}
		if v.Expr != nil {
			label.write(`\nExpr: `, `Expr`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Symbol != nil {
				t.down(nodeId, v.Symbol, `Symbol`)
			}
		}
		if v.Expr != nil {
			t.down(nodeId, v.Expr, `Expr`)
		}
	case *ast.Terminal:
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
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
		if v.Variable != nil {
			label.write(`\nVariable: `, `Variable`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Before != nil {
				t.down(nodeId, v.Before, `Before`)
			}
		}
		if v.Variable != nil {
			t.down(nodeId, v.Variable, `Variable`)
		}
	case *ast.Variable:
		name := types.Type_name[int32(v.Type)]
		label.write(`:\nType:`, name)
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.Keyword:
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
		}
		if v.Value != "" {
			label.write(`\nValue: \"`, v.Value, `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Before != nil {
				t.down(nodeId, v.Before, `Before`)
			}
		}
	case *ast.Space:
		for i, s := range v.Space {
			label.write(`\nSpace[`, strconv.Itoa(i), `]: \"`, s, `\"`)
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.NameExpr:
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if v.Name != nil {
			t.down(nodeId, v.Name, `Name`)
		}
		if v.AnyName != nil {
			t.down(nodeId, v.AnyName, `AnyName`)
		}
		if v.AnyNameExcept != nil {
			t.down(nodeId, v.AnyNameExcept, `AnyNameExcept`)
		}
		if v.NameChoice != nil {
			t.down(nodeId, v.NameChoice, `NameChoice`)
		}
	case *ast.Name:
		if v.Before != nil {
			label.write(`\nBefore: \"`, v.Before.String(), `\"`)
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
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
	case *ast.AnyName:
		if v.Underscore != nil {
			label.write(`\nUnderscore: `, v.Underscore.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Underscore != nil {
				t.down(nodeId, v.Underscore, `Underscore`)
			}
		}
	case *ast.AnyNameExcept:
		if v.Exclamation != nil {
			label.write(`\nExclamation: `, v.Exclamation.String())
		}
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Except != nil {
			label.write(`\nExcept: `, `Except`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.Exclamation != nil {
				t.down(nodeId, v.Exclamation, `Exclamation`)
			}
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.Except != nil {
			t.down(nodeId, v.Except, `Except`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	case *ast.NameChoice:
		if v.OpenParen != nil {
			label.write(`\nOpenParen: `, v.OpenParen.String())
		}
		if v.Left != nil {
			label.write(`\nLeft: `, `Left`)
		}
		if v.Pipe != nil {
			label.write(`\nPipe: `, v.Pipe.String())
		}
		if v.Right != nil {
			label.write(`\nRight: `, `Right`)
		}
		if v.CloseParen != nil {
			label.write(`\nCloseParen: `, v.CloseParen.String())
		}
		t.addNode(nodeId, map[string]string{attrLabel: label.finish()})
		if t.full {
			if v.OpenParen != nil {
				t.down(nodeId, v.OpenParen, `OpenParen`)
			}
		}
		if v.Left != nil {
			t.down(nodeId, v.Left, `Left`)

		}
		if t.full {
			if v.Pipe != nil {
				t.down(nodeId, v.Pipe, `Pipe`)
			}
		}
		if v.Right != nil {
			t.down(nodeId, v.Right, `Right`)
		}
		if t.full {
			if v.CloseParen != nil {
				t.down(nodeId, v.CloseParen, `CloseParen`)
			}
		}
	default:
		panic(fmt.Sprintf(`unknown ast node of type "%T" and value "%v"`, v, v))
	}
}

var attrLabel = string(gographviz.Label)

func (t *translator) down(nodeId string, to interface{}, edgeLabelName string) {
	nextNodeName := getTypeName(to)
	suffix := strconv.FormatUint(t.r.Uint64(), 10)
	t.addEdge(nodeId, nextNodeName+suffix, map[string]string{attrLabel: edgeLabelName})
	t.translate(to, nextNodeName, suffix)
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
