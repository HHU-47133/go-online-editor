package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	// "strings"
	// "container/heap"
)

const (
    N, M int = 250010, N << 2
    mod int = 1e9 + 7 // 998244353
)

var (
    in  *bufio.Reader = bufio.NewReader(os.Stdin)
    out *bufio.Writer = bufio.NewWriter(os.Stdout)
)

type ll = int64

/*******************************************************************************************************************/
type BIT struct { tr []int } 
func NewBIT() *BIT { return &BIT{ tr: make([]int, N) }}
func (b *BIT) sum(x int) int { res := 0; for x > 0 { res += b.tr[x]; x -= x & -x }; return res }
func (b *BIT) add(x, c int) { for x < N { b.tr[x] += c; x += x & -x }}
/*******************************************************************************************************************/
type hp struct { sort.IntSlice } 
func (h hp) Less(i, j int) bool { return h.IntSlice[i] < h.IntSlice[j] }
func (h *hp) Push(v interface{}) { h.IntSlice = append(h.IntSlice, v.(int)) }
func (h *hp) Pop() interface{} { a := h.IntSlice; v := a[len(a) - 1]; h.IntSlice = a[:len(a) - 1]; return v }
/*******************************************************************************************************************/
func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }
func gcd(a, b int) int { if b == 0 { return a }; return gcd(b, a % b) }
func lcm(a, b int) int { return (a * b / gcd(a, b)) }
func qmi(a, b, c int) int { r := 1; for ; b > 0; b >>= 1 { if b&1 == 1 { r = r * a % c }; a = a * a % c }; return r % c }
/*******************************************************************************************************************/
// var h [N]int; var w, e, ne [M]int; var idx int; func add(a, b, c int) { e[idx] = b; w[idx] = c; ne[idx] = h[a]; h[a] = idx; idx++ }

var n, m int

func solve() {
    fmt.Fscan(in, &n, &m)
    
}

func main() {
    defer out.Flush()
    var T int = 1
	// fmt.Fscan(in, &T)
    for ; T > 0; T-- { solve() }
}