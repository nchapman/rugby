import "os"
import "fmt"
import "math"
import "strconv"

class Body
  property x: Float
  property y: Float
  property z: Float
  property vx: Float
  property vy: Float
  property vz: Float
  property mass: Float

  def initialize(@x: Float, @y: Float, @z: Float, @vx: Float, @vy: Float, @vz: Float, @mass: Float)
  end
end

def energy(bodies: Array<Body>): Float
  e = 0.0
  i = 0
  while i < bodies.length
    body = bodies[i]
    e += 0.5 * body.mass * (body.vx*body.vx + body.vy*body.vy + body.vz*body.vz)
    j = i + 1
    while j < bodies.length
      body2 = bodies[j]
      dx = body.x - body2.x
      dy = body.y - body2.y
      dz = body.z - body2.z
      distance = math.Sqrt(dx*dx + dy*dy + dz*dz)
      e -= (body.mass * body2.mass) / distance
      j += 1
    end
    i += 1
  end
  e
end

def advance(bodies: Array<Body>, dt: Float)
  i = 0
  while i < bodies.length
    body = bodies[i]
    vx = body.vx
    vy = body.vy
    vz = body.vz
    j = i + 1
    while j < bodies.length
      body2 = bodies[j]
      dx = body.x - body2.x
      dy = body.y - body2.y
      dz = body.z - body2.z

      d_squared = dx*dx + dy*dy + dz*dz
      distance = math.Sqrt(d_squared)
      mag = dt / (d_squared * distance)

      vx -= dx * body2.mass * mag
      vy -= dy * body2.mass * mag
      vz -= dz * body2.mass * mag

      body2.vx += dx * body.mass * mag
      body2.vy += dy * body.mass * mag
      body2.vz += dz * body.mass * mag
      j += 1
    end
    body.vx = vx
    body.vy = vy
    body.vz = vz

    body.x += dt * vx
    body.y += dt * vy
    body.z += dt * vz
    i += 1
  end
end

def main
  n = 1000
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  solar_mass = 4.0 * math.Pi * math.Pi
  days_per_year = 365.24

  sun = Body.new(0.0, 0.0, 0.0, 0.0, 0.0, 0.0, solar_mass)

  jupiter = Body.new(
    4.84143144246472090,
    -1.16032004402742839,
    -0.103622044471123109,
    0.00166007664274403694 * days_per_year,
    0.00769901118419740425 * days_per_year,
    -0.0000690460016972063023 * days_per_year,
    0.000954791938424326609 * solar_mass
  )

  saturn = Body.new(
    8.34336671824457987,
    4.12479856412430479,
    -0.403523417114321381,
    -0.00276742510726862411 * days_per_year,
    0.00499852801234917238 * days_per_year,
    0.0000230417297573763929 * days_per_year,
    0.000285885980666130812 * solar_mass
  )

  uranus = Body.new(
    12.8943695621391310,
    -15.1111514016986312,
    -0.223307578892655734,
    0.00296460137564761618 * days_per_year,
    0.00237847173959480950 * days_per_year,
    -0.0000296589568540237556 * days_per_year,
    0.0000436624404335156298 * solar_mass
  )

  neptune = Body.new(
    15.3796971148509165,
    -25.9193146099879641,
    0.179258772950371181,
    0.00268067772490389322 * days_per_year,
    0.00162824170038242295 * days_per_year,
    -0.0000951592254519715870 * days_per_year,
    0.0000515138902046611451 * solar_mass
  )

  bodies = [sun, jupiter, saturn, uranus, neptune]

  # Offset momentum
  px = 0.0
  py = 0.0
  pz = 0.0
  i = 0
  while i < bodies.length
    body = bodies[i]
    px += body.vx * body.mass
    py += body.vy * body.mass
    pz += body.vz * body.mass
    i += 1
  end
  bodies[0].vx = -px / solar_mass
  bodies[0].vy = -py / solar_mass
  bodies[0].vz = -pz / solar_mass

  fmt.Printf("%.9f\n", energy(bodies))

  n.times -> {
    advance(bodies, 0.01)
  }

  fmt.Printf("%.9f\n", energy(bodies))
end
