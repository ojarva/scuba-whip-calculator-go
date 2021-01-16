Scuba transfer whip pressure calculator
=======================================

A small helper program for calculating pressures achieved when using a transfer whip between
scuba tanks, especially useful for source/destination twinsets with manifold that can be closed.

By default Van Der Waals equations are used for calculating amount of gas. Use `-use-ideal-gas` parameter to use ideal gas equation instead.

Installation
------------

There are no external requirements or dependencies.

Use `go build .` to build; then `./scuba-whip-calculator-go -h` to see instructions.

Example usage
-------------

```
./scuba-whip-calculator-go \
  -source-cylinder-volume 24 \
  -destination-cylinder-volume 17 \
  -destination-cylinder-pressure 80 \
  -source-cylinder-pressure 210 \
  -destination-cylinder-twinset \
  -source-cylinder-twinset
Equalizing with both manifolds closed
Source cylinders: 3432l, 143bar
Destination cylinders: 2968l, 175bar

Equalizing with destination manifold closed
Source cylinders: 3621l, 151bar
Destination cylinders: 2779l, 163bar

Equalizing with source manifold closed
Source cylinders: 3589l, 150bar
Destination cylinders: 2811l, 165bar

Equalizing with all manifolds open
Source cylinders: 3746l, 156bar
Destination cylinders: 2654l, 156bar

                               src bar  src l  dst bar  dst l improvement
         both manifolds closed     143   3432      175   2968      11.83%
   destination manifold closed     151   3621      163   2779       4.71%
        source manifold closed     150   3589      165   2811       5.91%
            all manifolds open     156   3746      156   2654       0.00%        
```

License
-------

BSD license, see LICENSE.md
