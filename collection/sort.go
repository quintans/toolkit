package collections

func Sort(array []interface{}, greater func(a, b interface{}) bool) {
	quickSort(greater, array, 0, len(array)-1)
}

func quickSort(greater func(a, b interface{}) bool, array []interface{}, low int, n int) {
	lo := low
	hi := n
	if lo < n {
		mid := array[(lo+hi)/2]
		for lo < hi {
			for lo < hi && greater(mid, array[lo]) {
				lo++
			}
			for lo < hi && greater(array[hi], mid) {
				hi--
			}
			if lo < hi {
				tmp := array[lo]
				array[lo] = array[hi]
				array[hi] = tmp
			}
		}
		if hi < lo {
			tmp := hi
			hi = lo
			lo = tmp
		}
		quickSort(greater, array, low, lo)
		var lw int
		if lo == low {
			lw = lo + 1
		} else {
			lw = lo
		}
		quickSort(greater, array, lw, n)
	}
}
