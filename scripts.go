package cardamomo

import (
  "fmt"
  "time"
  "strconv"
)

type Scripts struct {
  Functions map[int][]ScriptFunc
  TimersTimes []int
  FunctionsDatetimes map[int][]ScriptFunc
  TimersDatetimes []int
}

type ScriptFunc func () ()

func NewScripts() *Scripts {
  fmt.Printf("\n * Preparing scripts...\n\n")

  functions := make(map[int][]ScriptFunc, 0)
  timersTimes := make([]int, 0)
  functionsDatetimes := make(map[int][]ScriptFunc, 0)
  timersDatetimes := make([]int, 0)

  return &Scripts{Functions: functions, TimersTimes: timersTimes, FunctionsDatetimes: functionsDatetimes, TimersDatetimes: timersDatetimes}
}

// Adds a timer script

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
    for _, currentTimerTime := range s.TimersTimes {
      for _, function := range s.Functions[currentTimerTime] {
        function()
      }
    }

    for {
      select {
      case <- tricker.C:
        for _, currentTimerTime := range s.TimersTimes {
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

// Adds a concrete time script

func (s *Scripts) AddScriptAtTime(function ScriptFunc, datetime time.Time) {
  datetimeUnix := int(datetime.Unix())

  s.FunctionsDatetimes[datetimeUnix] = append(s.FunctionsDatetimes[datetimeUnix], function)

  if !SliceContains(s.TimersDatetimes, datetimeUnix) {
    startNewDatetimeTimer(s, datetimeUnix)
    s.TimersDatetimes = append(s.TimersDatetimes, datetimeUnix)
  }
}

func startNewDatetimeTimer(s *Scripts, datetime int) {
  tricker := time.NewTicker(time.Second)
  quit := make(chan struct{})
  go func() {
    currentDatetime := time.Now()
    currentDatetime, _ = time.Parse("3 04 PM", strconv.Itoa(currentDatetime.Hour()) + " " + strconv.Itoa(currentDatetime.Minute()) + " PM")

    if int(currentDatetime.Unix()) == datetime {
      for _, currentDatetime := range s.TimersDatetimes {
        for _, function := range s.FunctionsDatetimes[currentDatetime] {
          function()
        }
      }
    }

    for {
      select {
      case <- tricker.C:
        currentDatetime = time.Now()
        currentDatetime, _ = time.Parse("3 4 0 PM", strconv.Itoa(currentDatetime.Hour()) + " " + strconv.Itoa(currentDatetime.Minute()) + " " + strconv.Itoa(currentDatetime.Second()) + " PM")

        if int(currentDatetime.Unix()) == datetime {
          for _, currentDatetime := range s.TimersDatetimes {
            for _, function := range s.FunctionsDatetimes[currentDatetime] {
              function()
            }
          }
        }
      case <- quit:
        tricker.Stop()
        return
      }
    }
  }()
}
