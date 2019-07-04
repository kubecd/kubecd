/*
Copyright © 2019 Zedge, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

// json2yamlCmd represents the json2yaml command
var json2yamlCmd = &cobra.Command{
	Use:   "json2yaml",
	Short: "JSON to YAML conversion utility (stdin/stdout)",
	Long: ``,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var rawObject interface{}
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errors.Wrap(err, `reading stdin`)
		}
		err = json.Unmarshal(data, &rawObject)
		if err != nil {
			return errors.Wrap(err, `decoding JSON`)
		}
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		if err = encoder.Encode(&rawObject); err != nil {
			return errors.Wrap(err, `encoding YAML`)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(json2yamlCmd)
}
