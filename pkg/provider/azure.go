/*
 * Copyright 2018-2019 Zedge, Inc.
 * Copyright 2019-2020 Stig Sæther Nordahl Bakken
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
 *
 */

package provider

import (
	"github.com/kubecd/kubecd/pkg/model"
)

type AksClusterProvider struct{ baseClusterProvider }

func (p *AksClusterProvider) GetClusterInitCommands() ([][]string, error) {
	panic("implement me")
}

func (p *AksClusterProvider) GetClusterName() string {
	panic("implement me")
}

func (p *AksClusterProvider) GetUserName() string {
	panic("implement me")
}

func (p *AksClusterProvider) GetNamespace(env *model.Environment) string {
	panic("implement me")
}
