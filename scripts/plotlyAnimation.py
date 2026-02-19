import os
import math
import plotly.graph_objects as go
import networkx as nx

# ---------------- CONFIG ----------------
SNAPSHOT_DIR = "./snapshots"
NUM_FRAMES = 10         # how many snapshot files
SATS_PER_PLANE = 21
PLANE_START = 1
PLANE_END = 10
EARTH_RADIUS = 6371      # optional scale
OUTPUT_FILE = "orbit_animation.html"

plane_min = (PLANE_START - 1) * SATS_PER_PLANE + 1
plane_max = PLANE_END * SATS_PER_PLANE

# Cluster colors
colors = [
    "#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd",
    "#8c564b", "#e377c2", "#7f7f7f", "#bcbd22", "#17becf",
    "#393b79", "#637939", "#8c6d31", "#843c39", "#7b4173",
    "#3182bd", "#31a354", "#756bb1", "#636363", "#e6550d",
    "#fdae6b", "#6baed6", "#9ecae1", "#fd8d3c", "#fdd0a2",
    "#9e9ac8", "#bcbddc", "#e34a33", "#fc9272", "#dd1c77",
    "#f768a1", "#a6d854", "#ffd92f", "#1b9e77", "#d95f02",
    "#7570b3", "#e7298a", "#66a61e", "#e6ab02", "#a6761d"
]
# ---------------- LOAD CLUSTERS ----------------

def load_clusters(frame_id):
    cluster_file = f"./engLogs/clusters/{frame_id:03d}"
    cluster_graph = nx.Graph()
    cluster_heads = set()

    if not os.path.exists(cluster_file):
        return {}, set()

    with open(cluster_file, "r") as f:
        for line in f:
            line = line.strip()
            if "->" not in line:
                continue

            a, b = line.split("->")
            a, b = a.strip(), b.strip()

            cluster_graph.add_edge(a, b)

            if a == b:
                cluster_heads.add(a)

    components = list(nx.connected_components(cluster_graph))

    node_to_cluster = {}
    for idx, comp in enumerate(components):
        for node in comp:
            node_to_cluster[node] = idx

    return node_to_cluster, cluster_heads

# ---------------- LOAD SNAPSHOT ----------------
def load_snapshot(frame_id):
    positions = {}

    path = os.path.join(SNAPSHOT_DIR, f"{frame_id:03d}")
    print("Loading:",path)
    if not os.path.exists(path):
        return positions

    with open(path, "r") as f:
        for i, line in enumerate(f):
            parts = line.strip().split()
            if len(parts) != 3:
                continue

            sat_id = i + 1
            if not (plane_min <= sat_id <= plane_max):
                continue

            x, y, z = map(float, parts)
            positions[str(sat_id)] = (x, y, z)

    return positions

# ---------------- CREATE EARTH ----------------
def create_earth():
    u = [i * 2 * math.pi / 50 for i in range(50)]
    v = [i * math.pi / 50 for i in range(50)]

    x = []
    y = []
    z = []

    for ui in u:
        row_x = []
        row_y = []
        row_z = []
        for vi in v:
            row_x.append(EARTH_RADIUS * math.cos(ui) * math.sin(vi))
            row_y.append(EARTH_RADIUS * math.sin(ui) * math.sin(vi))
            row_z.append(EARTH_RADIUS * math.cos(vi))
        x.append(row_x)
        y.append(row_y)
        z.append(row_z)

    return go.Surface(x=x, y=y, z=z, opacity=0, showscale=False)

# ---------------- BUILD FRAMES ----------------
frames = []

for frame_id in range(NUM_FRAMES):
    positions = load_snapshot(frame_id)
    node_to_cluster, cluster_heads = load_clusters(frame_id)

    xs, ys, zs = [], [], []
    marker_colors = []

    for sat_id, (x, y, z) in positions.items():
        xs.append(x)
        ys.append(y)
        zs.append(z)

        if sat_id in cluster_heads:
            marker_colors.append("black")
        else:
            cluster_id = node_to_cluster.get(sat_id, -1)
            if cluster_id == -1:
                marker_colors.append("lightgray")
            else:
                marker_colors.append(
                    colors[cluster_id % len(colors)]
                )
    clusters = {}
    for node  in node_to_cluster:
        ch = node_to_cluster[node]
        if ch not in clusters:
            clusters[ch] = []
        clusters[ch].append(node)

    for cluster in clusters:
        if cluster == 0: continue
        if str(cluster) not in clusters[cluster]:
            # total += 1
            print(f"Clusterhead {cluster} doesnt include it self, {node_to_cluster.get(cluster,-1)}")
    frames.append(
        go.Frame(
            data=[go.Scatter3d(
                x=xs,
                y=ys,
                z=zs,
                mode="markers",
                marker=dict(size=5, color=marker_colors)
            )],
            traces=[1],
            name=str(frame_id)
        )
    )
# ---------------- INITIAL FIGURE ----------------
initial_positions = load_snapshot(0)
xs, ys, zs = [], [], []

for sat_id, (x, y, z) in initial_positions.items():
    xs.append(x)
    ys.append(y)
    zs.append(z)

satellites = go.Scatter3d(
    x=xs,
    y=ys,
    z=zs,
    mode="markers",
    marker=dict(size=4, color="white")
)

fig = go.Figure(
    data=[create_earth(), satellites],
    frames=frames
)

# ---------------- ANIMATION CONTROLS ----------------
fig.update_layout(
    scene=dict(
        xaxis=dict(visible=False),
        yaxis=dict(visible=False),
        zaxis=dict(visible=False),
        aspectmode="data"
    ),
    updatemenus=[{
    "type": "buttons",
    "buttons": [
        {
            "label": str(i),
            "method": "animate",
            "args": [
                [str(i)],
                dict(mode="immediate", frame=dict(duration=0, redraw=True))
            ]
        } for i in range(NUM_FRAMES)
    ]
}],
    sliders=[]  # remove slider if you just want buttons
)
   
# for cluster in node_to_cluster:
#     print(cluster)

# ---------------- SAVE ----------------
fig.write_html(OUTPUT_FILE)
print(f"Saved animation to {OUTPUT_FILE}")
