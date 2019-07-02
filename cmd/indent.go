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
package cmd

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
)

const defaultIndentLevel = 2
var indentLevel int

// indentCmd represents the indent command
var indentCmd = &cobra.Command{
	Use:   "indent file [file...]",
	Short: "canonically indent YAML files",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, file := range args {
			var rawObject yaml.Node
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return errors.Wrapf(err, `error reading %q`, file)
			}
			err = yaml.Unmarshal(data, &rawObject)
			if err != nil {
				return errors.Wrapf(err, `error decoding yaml in %q`, file)
			}
			tmpFile, err := ioutil.TempFile(path.Dir(file), path.Base(file)+"*")
			if err != nil {
				return errors.Wrapf(err, `error creating tmpfile for %q`, file)
			}
			//noinspection GoDeferInLoop
			defer func() { _ = os.Remove(tmpFile.Name()) }()
			encoder := yaml.NewEncoder(tmpFile)
			encoder.SetIndent(indentLevel)

			if err = encoder.Encode(&rawObject); err != nil {
				return errors.Wrapf(err, `error re-encoding `)
			}
			if err = os.Rename(tmpFile.Name(), file); err != nil {
				return errors.Wrapf(err, `error renaming %q to %q`, tmpFile.Name(), file)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indentCmd)
	indentCmd.Flags().IntVar(&indentLevel, "indent-level", defaultIndentLevel, "set indentation level")
}
