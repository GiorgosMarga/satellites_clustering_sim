import numpy as np
from astropy import units as u
from poliastro.bodies import Earth
from poliastro.twobody.orbit import Orbit

# -------------------------------
# Shell parameters
# -------------------------------
num_planes = 72
sats_per_plane = 21
altitude = 550 * u.km       # altitude above Earth's surface
inclination = 53 * u.deg
eccentricity = 0 * u.one    # circular orbits
arg_perigee = 0 * u.deg     # for circular orbits, usually 0
walker_f = 1                 # phasing parameter (Walker delta pattern)

# Semi-major axis
a = Earth.R + altitude

# Array to hold Orbit objects
orbits = []

# Spacing
raan_spacing = 360 / num_planes  # degrees between planes
ta_spacing = 360 / sats_per_plane  # degrees between satellites in a plane

# -------------------------------
# Generate satellites
# -------------------------------
for p in range(num_planes):
    raan = p * raan_spacing * u.deg
    
    for s in range(sats_per_plane):
        # Walker phasing: phase shift along plane
        true_anomaly = (s * ta_spacing + p * walker_f * ta_spacing / num_planes) * u.deg
        
        # Create orbit
        orbit = Orbit.from_classical(
            attractor=Earth,
            a=a,
            ecc=eccentricity,
            inc=inclination,
            raan=raan,
            argp=arg_perigee,
            nu=true_anomaly
        )
        
        # Add to array
        orbits.append(orbit)

# Check
print(f"Total satellites created: {len(orbits)}")
dt = 10 * u.min
for snapshot_id in range(50):
  for i, orb in enumerate(orbits):
      orbits[i] = orb.propagate(dt) 

  positions = np.array([orb.r.to(u.km).value for orb in orbits])
  with open(f"./snapshots/{snapshot_id:03d}", "+w") as f:
    for position in positions:
        f.write(f"{position[0]} {position[1]} {position[2]}\n")
    f.close()
  
          
