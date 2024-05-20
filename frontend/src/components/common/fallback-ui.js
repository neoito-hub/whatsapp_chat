import React from "react";
import MsgErrorIcon from "../../assets/img/msg-error.svg";

const FallbackUI = () => (
  <div className="max-w-5xl w-full float-left bg-white border border-primary px-4 py-6 md:p-6 rounded-2xl mb-4 mt-10 ml-10">
    <div className="w-full float-left flex flex-col items-center justify-center p-4 min-h-[360px] space-y-2.5">
      <img src={MsgErrorIcon} alt="Message Error" />
      <p className="text-gray-dark text-sm">Error Loading The Content</p>
      <button
        type="button"
        onClick={() => {
          // eslint-disable-next-line no-restricted-globals
          location.reload(true);
        }}
        className="flex-shrink-0 flex items-center text-sm text-purple h-8 px-3.5 rounded-md border border-[#6F42C1] shadow-xs font-semibold hover:text-white hover:bg-[#6F42C1] transition-colors"
      >
        Reload
      </button>
    </div>
  </div>
);

export default FallbackUI;
