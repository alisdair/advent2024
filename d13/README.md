# Day 13

With:

- $x_a$ and $x_b$ as the X movement of buttons A and B
- $y_a$ and $y_b$ as the Y movement of buttons A and B
- $x_p$ and $y_p$ as the prize location
- $c$ and $d$ as the number of times we press buttons A and B

Then these equations hold and must be solved:

```math
\begin{align}
c x_a + d x_b &= x_p \\
c y_a + d y_b &= y_p
\end{align}
```

Multiply first equation by $y_a$:

```math
c x_a y_a + d x_b y_a = x_p y_a
```

Multiply second equation by $x_a$:

```math
c x_a y_a + d x_a y_b = x_a y_p
```

Subtract to eliminate the first term and simplify:

```math
\begin{align}
d x_b y_a - d x_a y_b &= x_p y_a - x_a y_p \\
d (x_b y_a - x_a y_b) &= x_p y_a - x_a y_p \\
d &= \dfrac{x_p y_a - x_a y_p}{x_b y_a - x_a y_b}
\end{align}
```

Opposite approach for $c$:

```math
\begin{align}
c x_a y_b + d x_b y_b &= x_p y_b \\
c x_b y_a + d x_b y_b &= x_b y_p \\
c x_a y_b - c x_b y_a &= x_p y_b - x_b y_p \\
c (x_a y_b - x_b y_a) &= x_p y_b - x_b y_p \\
c &= \dfrac{x_p y_b - x_b y_p}{x_a y_b - x_b y_a}
\end{align}
```

### Example

```
Button A: X+94, Y+34
Button B: X+22, Y+67
Prize: X=8400, Y=5400
```

```math
\begin{align}
c &= \dfrac{x_p y_b - x_b y_p}{x_a y_b - x_b y_a} \\
  &= \dfrac{8400 \cdot 67 - 22 \cdot 5400}{94 \cdot 67 - 22 \cdot 34} \\
  &= \dfrac{562800 - 118800}{6298 - 748} \\
  &= \dfrac{444000}{5550} \\
  &= 80 \\
\end{align}
```

```math
\begin{align}
d &= \dfrac{x_p y_a - x_a y_p}{x_b y_a - x_a y_b} \\
  &= \dfrac{8400 \cdot 34 - 94 \cdot 5400}{22 \cdot 34 - 94 \cdot 67} \\
  &= \dfrac{285600 - 507600}{748 - 6298} \\
  &= \dfrac{-222000}{-5550} \\
  &= 40
\end{align}
```