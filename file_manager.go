package main

import (
  "log"
  "os"
  "path"
  "container/list"
  "time"
)

 const (
  _           = iota
  KB int64 = 1 << (10 * iota)
  MB
  GB
  TB
  WAITFOR = 10
  FRONT = "front"
  BACK = "back"
)

type empty struct{}

type FileManager struct {
  InputQueue chan string
  hasToSend chan empty
  OutputQueue chan string
  ReadyQueue chan string
  RemoveQueue chan string
  PutBackCh chan string
  queueSize int64
  qSet QueueSettings
  dirPath string
  fileNodes *list.List
}

type FileNode struct {
  Name string
  Size int64
}

func InitFileManager(dirPath string, qSet QueueSettings) *FileManager {
  fm := &FileManager{
    InputQueue: make(chan string),
    hasToSend: make(chan empty, 1),
    OutputQueue: make(chan string),
    PutBackCh: make(chan string),
    ReadyQueue: make(chan string),
    RemoveQueue: make(chan string),
    queueSize: 0,
    qSet: qSet,
    dirPath: dirPath,
    fileNodes: list.New(),
  }
  fm.Start()
  return fm
}

func (fm *FileManager) shrinkQueue() {
  frontPtr := fm.fileNodes.Back()
  for fm.queueSize > fm.qSet.MaxQueueSize && frontPtr != nil {
    fileNode := frontPtr.Value.(FileNode)
    prevPtr := frontPtr.Prev()

    if fm.qSet.RemoveSmallFiles || fileNode.Size >= fm.qSet.MinRemoveSize {
      log.Printf(
        "Queue size is %d, which is bigger than %d. Removing '%s', size %d\n",
        fm.queueSize, fm.qSet.MaxQueueSize, fileNode.Name, fileNode.Size,
      )
      fm.queueSize -= fileNode.Size
      fm.RemoveQueue <- fileNode.Name
      fm.fileNodes.Remove(frontPtr)
    } else {
      log.Printf("File '%s' is too small (%d < %d), skipping...\n",
        fileNode.Name, fileNode.Size, fm.qSet.MinRemoveSize,
      )
    }

    frontPtr = prevPtr
  }
}

func (fm *FileManager) putMark () {
  select {
    case fm.hasToSend <- empty{}:
      // do nothing
    default:
      // do nothing
  }
}

func (fm *FileManager) takeMark () {
  select {
    case <-fm.hasToSend:
      // do nothing
    default:
      // do nothing
  }
}

func (fm *FileManager) push(fn, direction string) {
  absPath := path.Join(fm.dirPath, fn)
  info, err := os.Stat(absPath)
  if err != nil {
    log.Printf("Couldn't get file info of '%s': %v\n", fn, err)
    fm.RemoveQueue <- fn
    return
  }

  if direction == FRONT {
    fm.fileNodes.PushFront(FileNode{
      Name: fn,
      Size: info.Size(),
    })
  } else {
    fm.fileNodes.PushBack(FileNode{
      Name: fn,
      Size: info.Size(),
    })
  }

  fm.queueSize += info.Size()
  log.Printf("Putting file '%s' [ %d bytes ] into the queue [ len=%d, %d bytes ]\n",
    fn, info.Size(), fm.fileNodes.Len(), fm.queueSize,
  )

  // remove backlogged files
  if fm.qSet.LimitQueueSize {
    fm.shrinkQueue()
  }

  if fm.fileNodes.Front() != nil {
    fm.putMark()
  } else {
    fm.takeMark()
  }

}

func (fm *FileManager) Start() {
  go func() {
    for {
      select {
      case fn := <-fm.InputQueue:
        fm.push(fn, FRONT)
      case fn := <-fm.PutBackCh:
        fm.push(fn, BACK)
      case <-fm.hasToSend:
        elemToSend := fm.fileNodes.Back()
        fileNode := elemToSend.Value.(FileNode)
        select {
        case fm.OutputQueue <- fileNode.Name:
          fm.queueSize -= fileNode.Size
          fm.fileNodes.Remove(elemToSend)
          if fm.fileNodes.Len() > 0 {
            fm.putMark()
          }
          log.Printf("Removed file '%s' [ %d bytes ] from the queue [ len=%d, %d bytes ]\n",
            fileNode.Name, fileNode.Size, fm.fileNodes.Len(), fm.queueSize,
          )
        default:
          fm.putMark()
          time.Sleep(WAITFOR * time.Millisecond)
        }
      }
    }
  }()
  go func() {
    for fn := range fm.RemoveQueue {
      absPath := path.Join(fm.dirPath, fn)
      err := os.Remove(absPath)
      if err != nil {
        log.Printf("Could not remove file '%s': %v\n", absPath, err)
        continue
      }
      log.Printf("Removed '%s'\n", absPath)
    }
  }()
}