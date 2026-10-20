## Math

Inline formulas should follow the prose baseline: the Pythagorean relation is $a^2 + b^2 = c^2$, a subscript is $x_i$, and Greek letters are $\alpha + \beta$.

The display formula checks a fraction and square root:

$$
x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}
$$

The sum checks upper and lower limits:

$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

The matrix checks row and column alignment:

$$
A = \begin{pmatrix} 1 & 2 \\ 3 & 4 \end{pmatrix}
$$

A macro defined in this document applies to later formulas: $\newcommand{\square}[1]{#1^2}\square{x}$, including the next paragraph.

$$
\left(\frac{\square{x}+1}{\sqrt{x}}\right)^2 + \left\langle x_i, y_i \right\rangle
$$

Delimiters inside a nested matrix should not split the outer cell:

$$
B = \begin{pmatrix} 1 & \begin{bmatrix}a & b\\c & d\end{bmatrix} \\ \frac12 & 4\end{pmatrix}
$$

Show only the scalable brace on the left:

$$
f(x)=\left\{\begin{matrix}x^2 & x>0\\0 & x\leq0\end{matrix}\right.
$$

Control: inline code `$x_i$` and dollar signs in the following code block should remain literal.

```text
$a^2 + b^2 = c^2$
$$this is not a formula to typeset$$
```

Ordinary text before and after formulas should wrap normally. An incomplete formula during streaming must not hide later content.

Unsupported commands should preserve their source: $\unknown{x}$, not be drawn as another formula.
