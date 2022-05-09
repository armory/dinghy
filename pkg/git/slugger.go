package git

type RepositoryInfo struct {
	Org  string
	Repo string
	Type string
}

// Slugger parses a URL from a Git provider webhook event and returns its constiuent parts.
//
//@&5!:..........................................      ....................:!P@@
//G^ ~5GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGBBBBBGGGGBY:    :Y55555555555555555Y?^ ~#
//: ?&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&#BGGGB#&&&&&B7    :5GGGGGGGGGGGGGGGGGGG~ ~
//  P&############################&BJ~:.   .:~75#&&&P^   .?GGGGGGGGGGGGGGGGGG? .
//  5&###########################&P:            :Y&#&#J.   !PGGGGGGGGGGGGGGGG? :
//  5&##########################&#^               5&##&G!   ^PGGGGGGGGGGGGGGG? :
//  5&########################&#P7                ~&###&&5:  :YGGGGGGGGGGGGGG? :
//  5&########################5!^::               ~######&#?   ?GGGGGGGGGGGGG? :
//  5&###########################&5              :#&&&&&&&&&G~  !PGGGGGGGGGGG? :
//  5&##########################&G:               !7777777?YBP   :!PGGGGGGGGG? :
//  5&############################BG.                        .     .?PGGGGGGG? :
//  5&##########################&#57.                                ~GGGGGGG? :
//  5&####&&&&&&###############&Y:                                    ~J5GGGG? :
//  5&###&G?~~?G&#############&J                                        .7GGG? :
//  5&####^    ^#############&P                                           !5G? :
//  5&###&5^::^5&############&7                                            ?G? :
//  P&####&###&&##############:                                          .?GG? .
//: 7&&&&&&&&&&&&&&&&&&&&&&&&#J??~                               ~777777?PGGP^ !
//#~ ^JPPPPPPPPPPPPPPPPPPPPPPPGGGP~                              75YYYYYYYJ7: 7&
//@@P7^...........................:.......................................:^?G@@
type Slugger interface {
	// NOTE: pushEvent should really be a stronger type than a generic map.
	Slug(pushEvent map[string]interface{}) (*RepositoryInfo, error)
}
