package cardamomo

import (
  "fmt"
  "time"
)

type Scripts struct {
  Functions map[int][]ScriptFunc
  TimersTimes []int
}

type ScriptFunc func () ()

func NewScripts() *Scripts {
  fmt.Printf("\n * Preparing scripts...\n\n")

  functions := make(map[int][]ScriptFunc, 0)
  timersTimes := make([]int, 0)

  return &Scripts{Functions: functions, TimersTimes: timersTimes}
}

func (s *Scripts) AddScript(function ScriptFunc, timerTime int) {
  s.Functions[timerTime] = append(s.Functions[timerTime], function)

  if !SliceContains(s.TimersTimes, timerTime) {
    startNewTimer(s, timerTime)
    s.TimersTimes = append(s.TimersTimes, timerTime)
  }
}

func startNewTimer(s *Scripts, timerTime int) {
  tricker := time.NewTicker(time.Duration(int32(timerTime)) * time.Second)
  quit := make(chan struct{})
  go func() {
    for {
      select {
      case <- tricker.C:
        for _, currentTimerTime := range s.TimersTimes {
          fmt.Printf("%s", currentTimerTime)
          for _, function := range s.Functions[currentTimerTime] {
    				function()
          }
        }
      case <- quit:
        tricker.Stop()
        return
      }
    }
  }()
}
