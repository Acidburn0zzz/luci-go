# Copyright 2018 The LUCI Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("go.chromium.org/luci/starlark/starlarkproto/testprotos/test.proto", "testprotos")
load("go.chromium.org/luci/starlark/starlarkproto/testprotos/proto2.proto", proto2="testprotos")

m = proto2.Proto2Message()

# Scalar field getter fails.
def try_getter():
  print(m.i)
assert.fails(try_getter, 'proto2 messages are not fully supported')

# Scalar field setter fails.
def try_setter():
  m.i = 123
assert.fails(try_setter, 'proto2 messages are not fully supported')

# Scalar field setter to a ptr value fails too.
def try_setter_ptr():
  m.i = testprotos.Simple()
assert.fails(try_setter_ptr, 'proto2 messages are not fully supported')

# Repeated fields work fine.
assert.eq(m.rep_i, [])
m.rep_i = [1, 2, 3]
assert.eq(m.rep_i, [1, 2, 3])

# Serialization also works.
assert.eq(proto.to_pbtext(m), "rep_i: 1\nrep_i: 2\nrep_i: 3\n")
