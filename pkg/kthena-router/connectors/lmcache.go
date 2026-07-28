/*
Copyright The Volcano Authors.

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

package connectors

// NewLMCacheConnector creates a new LMCache connector.
// vLLM's LMCacheConnectorV1 implements the generic KVConnectorBase_V1 contract:
// the prefill request returns kv_transfer_params, which must be forwarded on the
// decode request for the decode side to actually reuse the KV cache produced by
// prefill. This is the same contract NixlConnector already implements, so we
// reuse it here instead of falling back to the plain HTTP connector, which never
// reads the prefill response body and drops kv_transfer_params entirely.
func NewLMCacheConnector() KVConnector {
	return &NIXLConnector{
		name: "lmcache",
	}
}
