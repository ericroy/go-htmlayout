#include <stdlib.h>
#include <htmlayout.h>

// Main event function that dispatches to the appropriate event handler
BOOL CALLBACK ElementProc(LPVOID tag, HELEMENT he, UINT evtg, LPVOID prms)
{
	extern BOOL goElementProc(LPVOID, HELEMENT, UINT, LPVOID);
	return goElementProc(tag, he, evtg, prms);
}
LPELEMENT_EVENT_PROC ElementProcAddr = &ElementProc;

// Main event function that dispatches notify messages
LRESULT CALLBACK NotifyProc(UINT uMsg, WPARAM wParam, LPARAM lParam, LPVOID vParam)
{
	extern BOOL goNotifyProc(UINT, WPARAM, LPARAM, LPVOID);
	return goNotifyProc(uMsg, wParam, lParam, vParam);
}
LPHTMLAYOUT_NOTIFY NotifyProcAddr = &NotifyProc;

// Callback for results found during a select operation
BOOL CALLBACK SelectCallback(HELEMENT he, LPVOID param)
{
	extern BOOL goSelectCallback(HELEMENT, LPVOID);
	return goSelectCallback(he, param);
}
HTMLayoutElementCallback *SelectCallbackAddr = &SelectCallback;


INT ElementComparator(HELEMENT he1, HELEMENT he2, LPVOID pArg)
{
	extern INT goElementComparator(HELEMENT, HELEMENT, LPVOID);
	return goElementComparator(he1, he2, pArg);
}
ELEMENT_COMPARATOR *ElementComparatorAddr = (ELEMENT_COMPARATOR *)&ElementComparator;