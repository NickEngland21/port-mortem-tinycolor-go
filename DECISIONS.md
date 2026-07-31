# Decisions

## Value representation

The Go port stores channels as `float64` RGB values in the 0–255 range and
alpha in 0–1. This keeps conversion and HSL/HSV math close to the oracle while
making clamping explicit.

## API shape

Go constructors and functions use idiomatic exported names (`New`, `Mix`,
`Readability`, `MostReadable`) rather than reproducing JavaScript call syntax.
Mutating TinyColor-style methods use pointer receivers and return the receiver
to preserve chaining.

## Invalid input

Invalid input remains an invalid `Color` whose channels normalize to opaque
black, matching the observable oracle contract used by the fixture suite.

## Floating-point and rounding

Formatting rounds byte channels and alpha at the output boundary. HSV/HSL
conversion clamps unit values explicitly. The small floor step in
`Monochromatic` is retained because the oracle floors fractional HSV values to
four decimal places before conversion.

## Original tests

The upstream JavaScript test suite is not copied or modified here. The public
repository contains a 45-label Go-side adapter plus deterministic fixtures;
the original suite and kickoff hashes remain private evidence for honest
comparison. This is a deliberate publication boundary, not a claim of full
test-suite transplantation.

