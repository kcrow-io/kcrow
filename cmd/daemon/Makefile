include ../../Makefile.defs

TARGET := ../../bin/daemon

all: $(TARGET)

.PHONY: all $(TARGET)

$(TARGET): vendor ../../Makefile ../../Makefile.defs Makefile
	@$(ECHO_GO)
	$(GO_BUILD) -o $(TARGET)

vendor:
	$(QUIET) $(GO) mod vendor

clean:
	@$(ECHO_CLEAN)
	$(QUIET)rm -f $(TARGET)
	$(QUIET)$(GO_CLEAN)

install-bash-completion: $(TARGET) install-bash-completion-only

install-bash-completion-only:
	$(QUIET)$(INSTALL) -m 0755 -d $(DESTDIR_BASH_COMPLETION)
	$(QUIET)touch $(DESTDIR_BASH_COMPLETION)/$(TARGET).bash
	#./$(TARGET) completion bash > $(DESTDIR_BASH_COMPLETION)/$(TARGET).bash


install:
	$(QUIET)echo "install $(TARGET) to $(DESTDIR_BIN)"
	$(QUIET)$(INSTALL) -m 0755 -d $(DESTDIR_BIN)
	$(QUIET)$(INSTALL) -m 0755 $(TARGET) $(DESTDIR_BIN)
