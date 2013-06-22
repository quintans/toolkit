package toolkit

import (
	"bytes"
	"encoding/binary"
	r "reflect"
)

/**
* Collected methods which allow easy implementation of <code>hashCode</code>.
*
* Example use case:
* <pre>
*  public int hashCode(){
*    int result = HashCodeUtil.SEED;
*    //collect the contributions of various fields
*    result = HashCodeUtil.hash(result, fPrimitive);
*    result = HashCodeUtil.hash(result, fObject);
*    result = HashCodeUtil.hash(result, fArray);
*    return result;
*  }
* </pre>
 */

/**
 * An initial value for a <code>hashCode</code>, to which is added contributions
 * from fields. Using a non-zero value decreases collisons of <code>hashCode</code>
 * values.
 */
const HASH_SEED = 23
const prime_number = 37

/**
 * booleans.
 */
func HashBool(aSeed int, aBoolean bool) int {
	b := 0
	if aBoolean {
		b = 1
	}
	return firstTerm(aSeed) + b
}

func HashInt(aSeed int, aInt int) int {
	return firstTerm(aSeed) + aInt
}

func HashLong(aSeed int, aLong int64) int {
	return firstTerm(aSeed) + numberHashCode(aLong)
}

func HashDouble(aSeed int, aDouble float64) int {
	return firstTerm(aSeed) + numberHashCode(aDouble)
}

func HashString(aSeed int, aString string) int {
	return firstTerm(aSeed) + hashCode([]byte(aString))
}

func HashBytes(aSeed int, aBytes []byte) int {
	return firstTerm(aSeed) + hashCode(aBytes)
}

func HashBase(aSeed int, a Base) int {
	return firstTerm(aSeed) + a.HashCode()
}

func HashType(aSeed int, aType interface{}) int {
	typ := r.TypeOf(aType)
	if typ.Kind() == r.Ptr {
		typ = typ.Elem()
	}

	var t string
	if typ.PkgPath() != "" {
		t = typ.PkgPath() + "/" + typ.Name()
	} else {
		t = typ.Name()
	}

	return HashString(aSeed, t)
}

func Hash(aSeed int, aObject interface{}) int {
	result := aSeed
	if aObject == nil {
		result = HashInt(result, 0)
	} else {
		valuesVal := r.ValueOf(aObject)
		k := valuesVal.Kind()
		if k == r.Array || k == r.Slice {
			length := valuesVal.Len()
			for i := 1; i < length; i++ {
				item := valuesVal.Index(i).Interface()
				result = Hash(result, item)
			}
		} else if k == r.Bool {
			result = HashBool(result, aObject.(bool))
		} else if k >= r.Int && k <= r.Complex128 {
			result = HashInt(result, numberHashCode(aObject))
		} else if k == r.String {
			result = HashString(result, aObject.(string))
		} else if t, ok := aObject.(Base); ok {
			result = HashInt(result, t.HashCode())
		}
	}
	return result
}

func numberHashCode(aObject interface{}) int {
	h := 0
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, aObject)
	if err == nil {
		h = hashCode(buf.Bytes())
	}
	return h
}

func hashCode(what []byte) int {
	h := 0
	for _, v := range what {
		//h = 31*h + int(v)
		h = (h << 5) - h + int(v)
	}
	return h
}

func firstTerm(aSeed int) int {
	return prime_number * aSeed
}
