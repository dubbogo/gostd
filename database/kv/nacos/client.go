/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nacos

import (
	"sync"
)

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	clientPool           nacosClientPool
	configClientPool     nacosConfigClientPool
	clientPoolOnce       sync.Once
	configClientPoolOnce sync.Once
)

type nacosClientPool struct {
	sync.Mutex
	namingClient map[string]naming_client.INamingClient
}

type nacosConfigClientPool struct {
	sync.Mutex
	configClient map[string]config_client.IConfigClient
}

func initNacosClientPool() {
	clientPool.namingClient = make(map[string]naming_client.INamingClient)
}

func initNacosConfigClientPool() {
	configClientPool.configClient = make(map[string]config_client.IConfigClient)
}

// NewNacosNamingClient create nacos client
func NewNacosNamingClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (naming_client.INamingClient, error) {
	if !share {
		return newNamingClient(sc, cc)
	}
	clientPoolOnce.Do(initNacosClientPool)
	clientPool.Lock()
	defer clientPool.Unlock()
	if client, ok := clientPool.namingClient[name]; ok {
		return client, nil
	}

	client, err := newNamingClient(sc, cc)
	if err != nil {
		return nil, err
	}
	clientPool.namingClient[name] = client
	return client, err
}

// NewNacosConfigClient create config client
func NewNacosConfigClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (config_client.IConfigClient, error) {
	if !share {
		return newConfigClient(sc, cc)
	}
	configClientPoolOnce.Do(initNacosConfigClientPool)
	configClientPool.Lock()
	defer configClientPool.Unlock()
	if client, ok := configClientPool.configClient[name]; ok {
		return client, nil
	}

	client, err := newConfigClient(sc, cc)
	if err != nil {
		return nil, err
	}
	configClientPool.configClient[name] = client
	return client, err
}

func newNamingClient(sc []constant.ServerConfig, cc constant.ClientConfig) (naming_client.INamingClient, error) {
	cfg := vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc}
	return clients.NewNamingClient(cfg)
}

func newConfigClient(sc []constant.ServerConfig, cc constant.ClientConfig) (config_client.IConfigClient, error) {
	cfg := vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc}
	return clients.NewConfigClient(cfg)
}
