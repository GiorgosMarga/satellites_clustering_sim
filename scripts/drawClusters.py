import matplotlib.pyplot as plt
import networkx as nx
import math

# --- CONFIG ---
connections_file = "./engLogs/clusters/000"
positions_file = "./snapshots/000"
plot_width = 15
plot_height = 10
node_size = 50
LINE_ALPHA=0.1
PLANE_START = 1
PLANE_END = 5
SATS_PER_PLANE = 21
COLORS = [
    "#1f77b4",  # blue
    "#ff7f0e",  # orange
    "#2ca02c",  # green
    "#d62728",  # red
    "#9467bd",  # purple
    "#8c564b",  # brown
    "#e377c2",  # pink
    "#7f7f7f",  # gray
    "#bcbd22",  # olive
    "#17becf",  # cyan
    "#393b79",  # dark blue
    "#637939",  # dark olive
    "#8c6d31",  # mustard
    "#843c39",  # dark red
    "#7b4173",  # dark purple
]

# --- PARSE CONNECTIONS & CLUSTERS ---
connections = []
clusters = {}
node_ids = []

plane_min = (PLANE_START - 1) * SATS_PER_PLANE + 1
plane_max = PLANE_END * SATS_PER_PLANE



cluster_heads = set()
with open(connections_file, "r") as f:
    for line in f:
        line = line.strip()
        if not line or line.startswith("#"):
            continue

        if "->" in line:  # cluster relation
            a, b = line.split("->")
            a, b = a.strip(), b.strip()
            if plane_min <= int(a) <= plane_max and plane_min <= int(b) <= plane_max:
                # clusters.setdefault(a, []).append(b)
                if b not in clusters:
                    clusters[b] = []
                clusters[b].append(a)
                node_ids.extend([a, b])
                if a == b:
                    cluster_heads.add(a)

        elif "-" in line:  # connection
            a, b = line.split("-")
            a, b = a.strip(), b.strip()
            if plane_min <= int(a) <= plane_max and plane_min <= int(b) <= plane_max:
                connections.append((a, b))
                node_ids.extend([a, b])

node_ids = sorted(set(node_ids), key=lambda x: int(x))

# --- PARSE POSITIONS 3D ---
positions_3d = {}

with open(positions_file, "r") as f:
    for i, line in enumerate(f):
        parts = line.strip().split()
        if len(parts) != 3:
            continue

        sat_id = i + 1  # 1-based ID
        if not (plane_min <= sat_id <= plane_max):
            continue

        x, y, z = map(float, parts)
        positions_3d[str(sat_id)] = (x, y, z)

# --- PROJECT 3D -> 2D (Equirectangular) ---
positions_2d = {}

for node, (x, y, z) in positions_3d.items():
    r = math.sqrt(x**2 + y**2 + z**2)
    lat = math.asin(z / r)
    lon = math.atan2(y, x)

    x_2d = (lon + math.pi) / (2 * math.pi)
    y_2d = (lat + math.pi / 2) / math.pi

    positions_2d[node] = (x_2d, y_2d)

# --- CREATE GRAPH ---
G = nx.Graph()
G.add_edges_from(connections)

# --- BUILD CLUSTER GRAPH FOR COLORING ---
cluster_graph = nx.Graph()

for a, neighbors in clusters.items():
    for b in neighbors:
        cluster_graph.add_edge(a, b)

cluster_components = list(nx.connected_components(cluster_graph))

node_to_cluster = {}
for idx, comp in enumerate(cluster_components):
    for node in comp:
        node_to_cluster[node] = idx

node_colors = []
for node in G.nodes:
  for clusterId in clusters:
    for id in set(clusters[clusterId]):
      if id == node:
        node_colors.append(COLORS[int(clusterId) % len(COLORS)])
        

print(clusters)
# --- PLOT ---
plt.figure(figsize=(plot_width, plot_height))

nx.draw_networkx_nodes(
    G,
    positions_2d,
    node_size=node_size,
    node_color=node_colors,
    # cmap=plt.cm.tab20c,
)
# Draw edges with custom alpha
nx.draw_networkx_edges(
    G,
    positions_2d,
    alpha=LINE_ALPHA 
)
plt.title("Satellite Connections and Clusters (Flat Earth Projection)")
plt.axis("off")
plt.savefig("test.png")
plt.close()
