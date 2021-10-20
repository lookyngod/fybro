package mtproto

func (m *MTProto) ChannelsCreateChannel(broadcast,megagroup bool,title,about string) (*TL,error){
	return m.InvokeSync(TL_channels_createChannel{
		Broadcast:broadcast,
		Megagroup:megagroup,
		Title:title,
		About:about,
	})
}

func (m *MTProto) ChannelsDeleteChannel(channel TL) (*TL,error){
	return m.InvokeSync(TL_channels_deleteChannel{
		Channel:channel,
	})
}

func (m *MTProto) ChannelsKickFromChannel(channel TL,userId TL, kicked TL) (*TL,error){
	return m.InvokeSync(TL_channels_kickFromChannel{
		Channel:channel,
		User_id:userId,
		Kicked:kicked,
	})
}

func (m *MTProto) ChannelsEditAbout(channel TL, about string) (*TL,error){
	return m.InvokeSync(TL_channels_editAbout{
		Channel:channel,
		About:about,
	})
}

func (m *MTProto) ChannelsEditTitle(channel TL, title string) (*TL,error){
	return m.InvokeSync(TL_channels_editTitle{
		Channel:channel,
		Title:title,
	})
}

func (m *MTProto) ChannelsInviteToChannel(channel TL,users []TL) (*TL,error){
	return m.InvokeSync(TL_channels_inviteToChannel{
		Channel:channel,
		Users:users,
	})
}

func (m *MTProto) ChannelsToggleInvites(channel TL,enabled TL) (*TL,error){
	return m.InvokeSync(TL_channels_toggleInvites{
		Channel:channel,
		Enabled:enabled,
	})
}

func (m *MTProto) ChannelsToggleSignatures(channel TL,enabled TL) (*TL,error){
	return m.InvokeSync(TL_channels_toggleSignatures{
		Channel:channel,
		Enabled:enabled,
	})
}

func (m *MTProto) ChannelsEditPhoto(channel TL,photo TL) (*TL,error){
	return m.InvokeSync(TL_channels_editPhoto{
		Channel:channel,
		Photo:photo,
	})
}