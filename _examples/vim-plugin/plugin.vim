" Initialize the channel
if !exists('s:pianoVimJobId')
	let s:pianoVimJobId = 0
endif

let s:bin = 'piano-vim'

function! s:connect()
  let id = s:initRpc()
  
  if id == 0
    echoerr "pianoVim: cannot start rpc process"
  elseif id == -1
    echoerr "pianoVim: rpc process is not executable"
  else
    let s:pianoVimJobId = id 
  endif
endfunction

function! s:initRpc()
  if s:pianoVimJobId == 0
    return jobstart([s:bin], { 'rpc': v:true })
  return s:pianoVimJobId
endfunction

call s:connect()

