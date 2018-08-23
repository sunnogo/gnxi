/* Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gnmi

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/sunnogo/ygot/experimental/ygotutils"
	"github.com/sunnogo/ygot/ygot"

	pb "github.com/sunnogo/gnmi/proto/gnmi"
	cpb "github.com/sunnogo/go-genproto/googleapis/rpc/code"
)

// JSONUnmarshaler is the signature of the Unmarshal() function in the GoStruct code generated by openconfig ygot library.
type JSONUnmarshaler func([]byte, ygot.GoStruct) error

// GoStructEnumData is the data type to maintain GoStruct enum type.
type GoStructEnumData map[string]map[int64]ygot.EnumDefinition

// Model contains the model data and GoStruct information for the device to config.
type Model struct {
	modelData       []*pb.ModelData
	structRootType  reflect.Type
	schemaTreeRoot  *yang.Entry
	jsonUnmarshaler JSONUnmarshaler
	enumData        GoStructEnumData
}

// NewModel returns an instance of Model struct.
func NewModel(m []*pb.ModelData, t reflect.Type, r *yang.Entry, f JSONUnmarshaler, e GoStructEnumData) *Model {
	return &Model{
		modelData:       m,
		structRootType:  t,
		schemaTreeRoot:  r,
		jsonUnmarshaler: f,
		enumData:        e,
	}
}

// NewConfigStruct creates a ValidatedGoStruct of this model from jsonConfig. If jsonConfig is nil, creates an empty GoStruct.
func (m *Model) NewConfigStruct(jsonConfig []byte) (ygot.ValidatedGoStruct, error) {
	rootNode, stat := ygotutils.NewNode(m.structRootType, &pb.Path{})
	if stat.GetCode() != int32(cpb.Code_OK) {
		return nil, fmt.Errorf("cannot create root node: %v", stat)
	}

	rootStruct, ok := rootNode.(ygot.ValidatedGoStruct)
	if !ok {
		return nil, errors.New("root node is not a ygot.ValidatedGoStruct")
	}
	if jsonConfig != nil {
		if err := m.jsonUnmarshaler(jsonConfig, rootStruct); err != nil {
			return nil, err
		}
		if err := rootStruct.Validate(); err != nil {
			return nil, err
		}
	}
	return rootStruct, nil
}

// SupportedModels returns a list of supported models.
func (m *Model) SupportedModels() []string {
	mDesc := make([]string, len(m.modelData))
	for i, m := range m.modelData {
		mDesc[i] = fmt.Sprintf("%s %s", m.Name, m.Version)
	}
	sort.Strings(mDesc)
	return mDesc
}
