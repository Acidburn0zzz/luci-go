core.project(
    name = 'project',
    buildbucket = 'cr-buildbucket.appspot.com',
    scheduler = 'luci-scheduler.appspot.com',
    swarming = 'chromium-swarm.appspot.com',
)

core.recipe(
    name = 'noop',
    cipd_package = 'noop',
)

core.bucket(name = 'ci')

core.gitiles_poller(
    name = 'p1',
    bucket = 'ci',
    repo = 'https://noop.com',
    triggers = ['b1', 'b2', 'b3'],
)
core.gitiles_poller(
    name = 'p2',
    bucket = 'ci',
    repo = 'https://noop.com',
    triggers = ['b1', 'b2', 'b3'],
)

core.builder(
    name = 'b1',
    bucket = 'ci',
    recipe = 'noop',
    service_account = 'noop1@example.com',
    triggers = ['b2', 'b3'],
)
core.builder(
    name = 'b2',
    bucket = 'ci',
    recipe = 'noop',
    service_account = 'noop2@example.com',
    triggers = ['b3'],
)
core.builder(
    name = 'b3',
    bucket = 'ci',
    recipe = 'noop',
)

# Expect configs:
#
# === cr-buildbucket.cfg
# buckets: <
#   name: "ci"
#   acl_sets: "ci"
#   swarming: <
#     builders: <
#       name: "b1"
#       swarming_host: "chromium-swarm.appspot.com"
#       recipe: <
#         name: "noop"
#         cipd_package: "noop"
#         cipd_version: "refs/heads/master"
#       >
#       service_account: "noop1@example.com"
#     >
#     builders: <
#       name: "b2"
#       swarming_host: "chromium-swarm.appspot.com"
#       recipe: <
#         name: "noop"
#         cipd_package: "noop"
#         cipd_version: "refs/heads/master"
#       >
#       service_account: "noop2@example.com"
#     >
#     builders: <
#       name: "b3"
#       swarming_host: "chromium-swarm.appspot.com"
#       recipe: <
#         name: "noop"
#         cipd_package: "noop"
#         cipd_version: "refs/heads/master"
#       >
#     >
#   >
# >
# acl_sets: <
#   name: "ci"
# >
# ===
#
# === luci-scheduler.cfg
# job: <
#   id: "b1"
#   acl_sets: "ci"
#   buildbucket: <
#     server: "cr-buildbucket.appspot.com"
#     bucket: "ci"
#     builder: "b1"
#   >
# >
# job: <
#   id: "b2"
#   acls: <
#     role: TRIGGERER
#     granted_to: "noop1@example.com"
#   >
#   acl_sets: "ci"
#   buildbucket: <
#     server: "cr-buildbucket.appspot.com"
#     bucket: "ci"
#     builder: "b2"
#   >
# >
# job: <
#   id: "b3"
#   acls: <
#     role: TRIGGERER
#     granted_to: "noop1@example.com"
#   >
#   acls: <
#     role: TRIGGERER
#     granted_to: "noop2@example.com"
#   >
#   acl_sets: "ci"
#   buildbucket: <
#     server: "cr-buildbucket.appspot.com"
#     bucket: "ci"
#     builder: "b3"
#   >
# >
# trigger: <
#   id: "p1"
#   acl_sets: "ci"
#   triggers: "b1"
#   triggers: "b2"
#   triggers: "b3"
#   gitiles: <
#     repo: "https://noop.com"
#     refs: "refs/heads/master"
#   >
# >
# trigger: <
#   id: "p2"
#   acl_sets: "ci"
#   triggers: "b1"
#   triggers: "b2"
#   triggers: "b3"
#   gitiles: <
#     repo: "https://noop.com"
#     refs: "refs/heads/master"
#   >
# >
# acl_sets: <
#   name: "ci"
# >
# ===
#
# === project.cfg
# name: "project"
# ===
