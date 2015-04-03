package toolkit

import (
	"bytes"
	"encoding/binary"
	"reflect"
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

func HashUnit(aSeed int, aInt byte) int {
	return firstTerm(aSeed) + numberHashCode(aInt)
}

func HashTiny(aSeed int, aInt int8) int {
	return firstTerm(aSeed) + numberHashCode(aInt)
}

func HashShort(aSeed int, aInt int16) int {
	return firstTerm(aSeed) + numberHashCode(aInt)
}

func HashInteger(aSeed int, aInt int32) int {
	return firstTerm(aSeed) + numberHashCode(aInt)
}

func HashLong(aSeed int, aLong int64) int {
	return firstTerm(aSeed) + numberHashCode(aLong)
}

func HashFloat(aSeed int, aFloat float32) int {
	return firstTerm(aSeed) + numberHashCode(aFloat)
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
	typ := reflect.TypeOf(aType)
	if typ.Kind() == reflect.Ptr {
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
	result := hash(aSeed, aObject)
	if result == 0 {
		result = aSeed
	}
	return result
}

func hash(aSeed int, aObject interface{}) int {
	result := 0
	if aObject == nil {
		result = HashInt(aSeed, 0)
	} else if t, ok := aObject.(Hasher); ok {
		result = HashInt(aSeed, t.HashCode())
	} else {
		v := reflect.ValueOf(aObject)
		k := v.Kind()
		if k == reflect.Array || k == reflect.Slice {
			length := v.Len()
			for i := 1; i < length; i++ {
				item := v.Index(i).Interface()
				result = Hash(aSeed, item)
			}
		} else if k == reflect.Bool {
			result = HashBool(aSeed, aObject.(bool))
		} else if k >= reflect.Int && k <= reflect.Complex128 {
			result = HashInt(aSeed, numberHashCode(aObject))
		} else if k == reflect.String {
			result = HashString(aSeed, aObject.(string))
		} else if k == reflect.Ptr {
			// tries pointer element
			o := v.Elem().Interface()
			r := hash(aSeed, o)
			if r == 0 {
				// no luck with the pointer. lets use pointer address value
				r = HashInt(aSeed, int(v.Pointer()))
			}
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
