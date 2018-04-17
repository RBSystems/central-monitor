package elk

const ControlProcsAndSuppressedRooms = `
{
	"_source": false,
				
	"query": {
		"query_string": {
			"query": "( _type:room AND notifications-suppressed:true ) OR ( _type:control-processor AND NOT notifications_suppressed:true )"
		}
	},
	"size": 10000
}
`
