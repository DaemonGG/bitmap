// An implementation of an arbitrary length bitmap
// example:
// 	 btmap64 := New(129)  // initialize a bitmap length 129
// 	 btmap64.Set(0, 3)    // set bits on index 0~3
// 	 fmt.Println(btmap64)
// 	 btmap64.Set(63, 7)
// 	 fmt.Println(btmap64)
// 	 succeed_ := btmap64.Set(64, 65)  // Set bits should fail
// 	 fmt.Println(btmap64, succeed_)

package bitmap

import (
  "log"
  "fmt"
)

type BitMap struct {
  bitmap []uint64
  length int
  num_set int
  num_one_sections int
}

func New(max int) *BitMap {
  return &BitMap{
    make([]uint64, (max + 63)/64),
    max,
    0,
    0}
}

const (
  unitStart = 0
  unitEnd = 64
)

func (this *BitMap) Len() int {
  return this.length
}

func (this *BitMap) NumSet() int {
  return this.num_set
}

func (this *BitMap) Set(start int, span int) bool {
  if this.setInternal(start, span, true) {
    this.setInternal(start, span, false)
    this.num_one_sections ++
    return true
  }
  return false
}

func (this *BitMap) setInternal(start int, span int, dry bool) bool {
  if start < 0 || start + span > this.length {
    log.Fatalf("Wrong parameters. Setting bits beyond boundary." +
                "length: %d, start: %d, span: %d\n", this.length, start, span)
  }
  start_idx := start/64
  end_idx := (start + span - 1)/64
  if start_idx == end_idx {
    if setUnit(
      &this.bitmap[start_idx], start % 64, (start + span - 1) % 64 + 1, !dry) {

      if !dry {
        this.num_set += span
      }
      return true
    } else {
      return false
    }
  }
  if setUnit(&this.bitmap[start_idx], start % 64, unitEnd, !dry) == true &&
     setUnit(&this.bitmap[end_idx], unitStart, (start + span - 1) % 64 + 1, !dry) == true {
     for i := start_idx + 1; i < end_idx; i += 1 {
       if !setUnit(&this.bitmap[i], unitStart, unitEnd, !dry) {
         return false
       }
     }
     if !dry {
       this.num_set += span
     }
     return true
  }
  return false
}

// This unit is only set when there is no contention.
func setUnit(origin *uint64, start int, end int, commit bool) bool {
  copy_bits := *origin
  var a, b uint64
  a = 1 << uint(start) - 1
  b = 1 << uint(end) - 1
  mask := a ^ b
  if copy_bits & mask != 0 {
    return false
  }

  if commit {
    *origin = copy_bits | mask
  }
  return true
}

func (this *BitMap) String() string {
  area_start := -1
  area_end := -1
  log := ""
  for i := 0; i < this.length; i += 1 {
    unit := this.bitmap[i/64]
    is_set := (1 << uint(i%64)) & unit
    if is_set != 0 {
      if area_start < 0 {
        area_start = i
        area_end = -1
      }
    } else {
        if area_start >= 0 && area_end < 0 {
          area_end = i
          log += fmt.Sprintf("[%d, %d)", area_start, area_end)
          area_start = -1
          area_end = -1
        }
    }
  }
  if area_start >= 0 && area_end < 0 {
    log += fmt.Sprintf("[%d, %d)", area_start, this.length)
  }
  return fmt.Sprintf(
           "Length: %d, Len of array: %d," +
           "# set slots: %d\nOccupied areas: %s\n",
           this.length, len(this.bitmap), this.num_set, log)
}

func (this *BitMap) NumSections() (int, int) {
  res := 0
  num_one := 0
  prev_set := true
  for i := 0; i < this.length; i += 1 {
    unit := this.bitmap[i/64]
    is_set := (1 << uint(i%64)) & unit
    if is_set == 0 {
      if prev_set {
        res += 1
      }
      prev_set = false
    } else {
      if !prev_set {
        num_one += 1
      }
      prev_set = true
    }
  }
  if num_one > this.num_one_sections {
    log.Fatalf("Counting # of one sections wrong! %d Vs. %d\n%v\n",
       num_one, this.num_one_sections, this)
  }
  return res, this.num_one_sections
}

func test() {
  btmap64 := New(129)

  btmap64.Set(0, 3)
  fmt.Println(btmap64)
  btmap64.Set(3, 61)
  fmt.Println(btmap64)
  btmap64.Set(63, 7)
  fmt.Println(btmap64)
  succeed_ := btmap64.Set(64, 65)
  fmt.Println(btmap64, succeed_)
}
