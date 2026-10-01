## 数学公式

行内公式应跟随正文基线：勾股关系为 $a^2 + b^2 = c^2$，变量带有下标 $x_i$，希腊字母为 $\alpha + \beta$。

独立公式检查分式和根号：

$$
x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}
$$

求和检查上下限：

$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

矩阵检查行列对齐：

$$
A = \begin{pmatrix} 1 & 2 \\ 3 & 4 \end{pmatrix}
$$

宏定义在同篇文档的后续公式中生效：$\newcommand{\square}[1]{#1^2}\square{x}$，下一段继续使用。

$$
\left(\frac{\square{x}+1}{\sqrt{x}}\right)^2 + \left\langle x_i, y_i \right\rangle
$$

嵌套矩阵的内层分隔符不应拆开外层单元格：

$$
B = \begin{pmatrix} 1 & \begin{bmatrix}a & b\\c & d\end{bmatrix} \\ \frac12 & 4\end{pmatrix}
$$

只显示左侧伸缩大括号：

$$
f(x)=\left\{\begin{matrix}x^2 & x>0\\0 & x\leq0\end{matrix}\right.
$$

对照：行内代码 `$x_i$` 和下面代码块里的美元符号应保持原样。

```text
$a^2 + b^2 = c^2$
$$这不是需要排版的公式$$
```

公式前后的普通文字应继续正常换行，流式输出时未闭合的公式不应导致后续内容消失。

不支持的命令应保留源码：$\unknown{x}$，不要把它误画成已有公式。
